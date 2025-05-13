package istio

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/syedomair/istio-config-manager/config"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	istioclientset "istio.io/client-go/pkg/clientset/versioned"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientnetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

func ApplyHeaderOverride(client istioclientset.Interface, cfg config.AppConfig) error {
	vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

	vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VirtualService: %v", err)
	}

	log.Println("Original VirtualService fetched. Checking for existing header+URI match...")

	targetHeader := "x-user-type"
	targetValue := "beta1"
	targetPrefix := "/user/users"
	targetRewrite := "/v2/users"
	targetSubset := cfg.SubsetName // e.g., v2

	// Check if an identical match+rewrite+route already exists
	for _, httpRoute := range vs.Spec.Http {
		// Check each match
		for _, match := range httpRoute.Match {
			hdr, hdrOk := match.Headers[targetHeader]
			uri := match.Uri.GetPrefix()
			if hdrOk && hdr.GetExact() == targetValue && uri == targetPrefix {
				// Match exists, now check rewrite and destination
				if httpRoute.Rewrite != nil && httpRoute.Rewrite.Uri == targetRewrite {
					for _, r := range httpRoute.Route {
						if r.Destination != nil && r.Destination.Host == cfg.Host && r.Destination.Subset == targetSubset {
							log.Println("Matching header-based route already exists. Skipping update.")
							return nil
						}
					}
				}
			}
		}
	}

	log.Println("No matching route found. Adding new header-based route.")

	// Define new route
	newMatch := &networkingv1beta1.HTTPMatchRequest{
		Headers: map[string]*networkingv1beta1.StringMatch{
			targetHeader: {MatchType: &networkingv1beta1.StringMatch_Exact{Exact: targetValue}},
		},
		Uri: &networkingv1beta1.StringMatch{
			MatchType: &networkingv1beta1.StringMatch_Prefix{
				Prefix: targetPrefix,
			},
		},
	}

	newRoute := networkingv1beta1.HTTPRoute{
		Match: []*networkingv1beta1.HTTPMatchRequest{newMatch},
		Rewrite: &networkingv1beta1.HTTPRewrite{
			Uri: targetRewrite,
		},
		Route: []*networkingv1beta1.HTTPRouteDestination{
			{
				Destination: &networkingv1beta1.Destination{
					Host:   cfg.Host,
					Subset: targetSubset,
				},
			},
		},
	}

	// Prepend the new route (higher priority)
	vs.Spec.Http = append([]*networkingv1beta1.HTTPRoute{&newRoute}, vs.Spec.Http...)

	_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update VirtualService: %v", err)
	}

	log.Printf("VirtualService %s updated with header-based routing.\n", cfg.VirtualServiceName)
	return nil
}

func ApplyTrafficSplit(client istioclientset.Interface, cfg config.AppConfig) error {
	vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

	vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VirtualService: %v", err)
	}

	log.Println("Original VirtualService fetched. Checking for existing traffic split route...")

	found := false

	v1Weight, err := strconv.ParseInt(cfg.TrafficWeight1, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid integer value given for trafiic weight 1: %v", err)
	}
	v1Weight32 := int32(v1Weight)

	v2Weight, err := strconv.ParseInt(cfg.TrafficWeight2, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid integer value given for trafiic weight 2: %v", err)
	}
	v2Weight32 := int32(v2Weight)

	for _, httpRoute := range vs.Spec.Http {
		// Match route with 2 destinations for same host with v1 and v2 subsets
		if len(httpRoute.Route) == 2 &&
			httpRoute.Route[0].Destination.Host == cfg.Host &&
			httpRoute.Route[1].Destination.Host == cfg.Host {

			subset1 := httpRoute.Route[0].Destination.Subset
			subset2 := httpRoute.Route[1].Destination.Subset

			if (subset1 == "v1" && subset2 == "v2") || (subset1 == "v2" && subset2 == "v1") {
				// Update weights
				httpRoute.Route[0].Weight = v1Weight32
				httpRoute.Route[1].Weight = v2Weight32
				found = true
				log.Println("Existing traffic split rule found. Updating weights...")
				break
			}
		}
	}

	if !found {
		log.Println("No existing traffic split found. Adding new A/B routing rule...")

		newSplitRoute := &networkingv1beta1.HTTPRoute{
			Route: []*networkingv1beta1.HTTPRouteDestination{
				{
					Destination: &networkingv1beta1.Destination{
						Host:   cfg.Host,
						Subset: "v1",
					},
					Weight: v1Weight32,
				},
				{
					Destination: &networkingv1beta1.Destination{
						Host:   cfg.Host,
						Subset: "v2",
					},
					Weight: v2Weight32,
				},
			},
		}

		// Append to existing routes
		vs.Spec.Http = append(vs.Spec.Http, newSplitRoute)
	}

	_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update VirtualService with traffic split: %v", err)
	}

	log.Printf("Updated VirtualService %s with v1:%d%%, v2:%d%% traffic split\n",
		cfg.VirtualServiceName, v1Weight, v2Weight)
	return nil
}
func ApplyDelayFault(client istioclientset.Interface, cfg config.AppConfig) error {
	vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

	vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
	if err != nil {
		return err
	}

	faultDelayMillis, err := strconv.ParseInt(cfg.FaultDelayMillis, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid integer value given for FaultDelayMillis : %v", err)
	}

	faultMatch := &networkingv1beta1.HTTPMatchRequest{
		Headers: map[string]*networkingv1beta1.StringMatch{
			"x-user-type": {
				MatchType: &networkingv1beta1.StringMatch_Exact{Exact: "test"},
			},
		},
	}

	faultRoute := &networkingv1beta1.HTTPRoute{
		Match: []*networkingv1beta1.HTTPMatchRequest{faultMatch},
		Fault: &networkingv1beta1.HTTPFaultInjection{
			Delay: &networkingv1beta1.HTTPFaultInjection_Delay{
				Percent: 100,
				HttpDelayType: &networkingv1beta1.HTTPFaultInjection_Delay_FixedDelay{
					FixedDelay: durationpb.New(time.Millisecond * time.Duration(faultDelayMillis)),
				},
			},
		},
		Route: []*networkingv1beta1.HTTPRouteDestination{
			{
				Destination: &networkingv1beta1.Destination{
					Host:   cfg.Host,
					Subset: cfg.SubsetName,
				},
			},
		},
	}

	vs.Spec.Http = append([]*networkingv1beta1.HTTPRoute{faultRoute}, vs.Spec.Http...)

	_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
	log.Printf("Updated VirtualService %s with with fault\n", cfg.VirtualServiceName)
	return err
}

func CreateTrafficMirroring(client istioclientset.Interface, cfg config.AppConfig) error {

	vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

	vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
	if err != nil {
		return err
	}

	mirrorRoute := &networkingv1beta1.HTTPRoute{
		Route: []*networkingv1beta1.HTTPRouteDestination{
			{
				Destination: &networkingv1beta1.Destination{
					Host:   cfg.Host,
					Subset: cfg.SubsetName,
				},
			},
		},
		Mirror: &networkingv1beta1.Destination{
			Host:   cfg.Host,
			Subset: cfg.Destination,
		},
	}

	vs.Spec.Http = append([]*networkingv1beta1.HTTPRoute{mirrorRoute}, vs.Spec.Http...)
	_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
	log.Printf("Updated VirtualService %s with with mirror\n", cfg.VirtualServiceName)
	return err
}

func ApplyTimeout(client istioclientset.Interface, cfg config.AppConfig) error {

	vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

	vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
	if err != nil {
		return err
	}

	timeoutDuration, err := strconv.ParseInt(cfg.TimeoutDuration, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid integer value given for trafiic weight 1: %v", err)
	}

	for _, httpRoute := range vs.Spec.Http {
		if len(httpRoute.Match) > 0 && httpRoute.Match[0].Uri != nil && httpRoute.Match[0].Uri.GetPrefix() == "/user/users" {
			log.Printf("Updating timeout for route to %v\n", timeoutDuration)
			httpRoute.Timeout = durationpb.New(time.Duration(timeoutDuration) * time.Second)
		}
	}

	_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update VirtualService with timeout: %v", err)
	}

	log.Printf("Timeout set to %v for VirtualService %s.\n", timeoutDuration, cfg.VirtualServiceName)
	return nil
}

func ApplyRetryPolicy(client istioclientset.Interface, cfg config.AppConfig) error {

	vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

	vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
	if err != nil {
		return err
	}

	retryAttempt, err := strconv.ParseInt(cfg.RetryAttempt, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid integer value given for RetryAttempt : %v", err)
	}
	retryAttempt32 := int32(retryAttempt)

	retryPerTryTimeout, err := strconv.ParseInt(cfg.RetryPerTryTimeout, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid integer value given for retryPerTryTimeout : %v", err)
	}

	retryPerTryTimeout32 := int32(retryPerTryTimeout)

	for _, httpRoute := range vs.Spec.Http {
		if len(httpRoute.Match) > 0 && httpRoute.Match[0].Uri != nil && httpRoute.Match[0].Uri.GetPrefix() == "/user/users" {
			log.Printf("Updating retry policy to \n")

			httpRoute.Retries = &networkingv1beta1.HTTPRetry{
				Attempts:      retryAttempt32,
				PerTryTimeout: durationpb.New(time.Duration(retryPerTryTimeout32) * time.Second),
				RetryOn:       cfg.RetryOn,
			}
		}
	}

	_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update VirtualService with timeout: %v", err)
	}

	log.Printf("Retries policy set for VirtualService %s.\n", cfg.VirtualServiceName)
	return nil
}

/*
func ApplyCircuitBreaker(client istioclientset.Interface, cfg config.AppConfig) error {

		drClient := client.NetworkingV1beta1().DestinationRules(cfg.Namespace)

		maxEjectionPercent := 100
		maxEjectionPercent32 := int32(maxEjectionPercent)
		consecutiveGatewayErrors := 1
		consecutiveGatewayErrors32 := int32(consecutiveGatewayErrors)

		dr := &networkingv1beta1.DestinationRule{
			Host: cfg.Host,
			TrafficPolicy: &networkingv1beta1.TrafficPolicy{
				ConnectionPool: &networkingv1beta1.ConnectionPoolSettings{
					Tcp: &networkingv1beta1.ConnectionPoolSettings_TCPSettings{
						MaxConnections: 1,
					},
					Http: &networkingv1beta1.ConnectionPoolSettings_HTTPSettings{
						Http1MaxPendingRequests:  1,
						MaxRequestsPerConnection: 1,
					},
				},
				OutlierDetection: &networkingv1beta1.OutlierDetection{
					ConsecutiveGatewayErrors: wrapperspb.UInt32(uint32(consecutiveGatewayErrors32)),
					Interval:                 prototime("5s"),
					BaseEjectionTime:         prototime("30s"),
					MaxEjectionPercent:       maxEjectionPercent32,
				},
			},
		}



		_, err := drClient.Create(context.Background(), dr, metaV1.CreateOptions{})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				log.Printf("DestinationRule %s already exists. Skipping creation.\n", cfg.Destination)
				return nil
			}
			return fmt.Errorf("failed to create DestinationRule: %v", err)
		}

		log.Printf("DestinationRule %s created with circuit breaker settings.\n", cfg.Destination)
		return nil
	}

	func prototime(duration string) *durationpb.Duration {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return nil
		}
		return durationpb.New(d)
	}
*/
func ApplyCircuitBreaker(client istioclientset.Interface, cfg config.AppConfig) error {
	drClient := client.NetworkingV1beta1().DestinationRules(cfg.Namespace)

	maxEjectionPercent := 100
	consecutiveGatewayErrors := 1

	// Create the API version of DestinationRule
	apiDr := &networkingv1beta1.DestinationRule{
		Host: cfg.Host,
		TrafficPolicy: &networkingv1beta1.TrafficPolicy{
			ConnectionPool: &networkingv1beta1.ConnectionPoolSettings{
				Tcp: &networkingv1beta1.ConnectionPoolSettings_TCPSettings{
					MaxConnections: 1,
				},
				Http: &networkingv1beta1.ConnectionPoolSettings_HTTPSettings{
					Http1MaxPendingRequests:  1,
					MaxRequestsPerConnection: 1,
				},
			},
			OutlierDetection: &networkingv1beta1.OutlierDetection{
				ConsecutiveGatewayErrors: wrapperspb.UInt32(uint32(consecutiveGatewayErrors)),
				Interval:                 prototime("5s"),
				BaseEjectionTime:         prototime("30s"),
				MaxEjectionPercent:       int32(maxEjectionPercent),
			},
		},
	}

	// Convert to the client-go version
	clientDr := &clientnetworkingv1beta1.DestinationRule{
		ObjectMeta: metaV1.ObjectMeta{
			Name: cfg.Destination, // Make sure to set the name
		},
		Spec: *apiDr,
	}

	_, err := drClient.Create(context.Background(), clientDr, metaV1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			log.Printf("DestinationRule %s already exists. Skipping creation.\n", cfg.Destination)
			return nil
		}
		return fmt.Errorf("failed to create DestinationRule: %v", err)
	}

	log.Printf("DestinationRule %s created with circuit breaker settings.\n", cfg.Destination)
	return nil
}

func prototime(duration string) *durationpb.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil
	}
	return durationpb.New(d)
}
