package istio

import (
	"context"
	"fmt"
	"log"

	"github.com/syedomair/istio-config-manager/config"
	"github.com/syedomair/istio-config-manager/kube"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	v1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istioclientset "istio.io/client-go/pkg/clientset/versioned"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ApplyConfig(clients *kube.Clients, cfg config.AppConfig) error {

	if !cfg.SkipVirtualService {
		vs := &v1beta1.VirtualService{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      cfg.VirtualServiceName,
				Namespace: cfg.Namespace,
			},
			Spec: networkingv1beta1.VirtualService{
				Hosts:    []string{cfg.Host},
				Gateways: []string{"mesh"},
				Http: []*networkingv1beta1.HTTPRoute{
					{
						Route: []*networkingv1beta1.HTTPRouteDestination{
							{
								Destination: &networkingv1beta1.Destination{
									Host: cfg.Destination,
								},
							},
						},
					},
				},
			},
		}

		vsClient := clients.Istio.NetworkingV1beta1().VirtualServices(cfg.Namespace)
		_, err := vsClient.Create(context.Background(), vs, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create VirtualService: %w", err)
		}
		fmt.Printf("VirtualService %s created in namespace %s\n", cfg.VirtualServiceName, cfg.Namespace)
	} else {
		fmt.Println("Skipping VirtualService creation (SKIP_VIRTUAL_SERVICE is set)")
	}

	if !cfg.SkipDestinationRule {
		dr := v1beta1.DestinationRule{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      cfg.VirtualServiceName + "-dr",
				Namespace: cfg.Namespace,
			},
			Spec: networkingv1beta1.DestinationRule{
				Host: cfg.Destination,
				Subsets: []*networkingv1beta1.Subset{
					{
						Name: cfg.SubsetName,
						Labels: map[string]string{
							"version": cfg.SubsetName,
						},
					},
				},
			},
		}

		_, err := clients.Istio.NetworkingV1beta1().DestinationRules(cfg.Namespace).
			Create(context.Background(), &dr, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create DestinationRule: %w", err)
		}
		fmt.Printf("DestinationRule %s created in namespace %s\n", dr.Name, cfg.Namespace)
	} else {
		fmt.Println("Skipping DestinationRule creation (SKIP_DESTINATION_RULE is set)")
	}

	return nil
}

/*
	func ApplyHeaderOverride(client istioclientset.Interface, cfg config.AppConfig) error {
		vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

		// Get existing VirtualService
		vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get VirtualService: %v", err)
		}

		log.Println("Original VirtualService fetched. Updating match rules...")

		// Add or override match rule to route only requests with header x-user-type: beta
		newMatch := &networkingv1beta1.HTTPMatchRequest{
			Headers: map[string]*networkingv1beta1.StringMatch{
				"x-user-type": {MatchType: &networkingv1beta1.StringMatch_Exact{Exact: "beta"}},
			},
			Uri: &networkingv1beta1.StringMatch{
				MatchType: &networkingv1beta1.StringMatch_Prefix{
					Prefix: "/user/users",
				},
			},
		}

		newRoute := networkingv1beta1.HTTPRoute{
			Match: []*networkingv1beta1.HTTPMatchRequest{newMatch},
			Rewrite: &networkingv1beta1.HTTPRewrite{
				Uri: "/v2/users"},
			Route: []*networkingv1beta1.HTTPRouteDestination{
				{
					Destination: &networkingv1beta1.Destination{
						Host:   cfg.Host,
						Subset: cfg.SubsetName, // e.g., v2
					},
				},
			},
		}

		// Prepend new match rule to HTTP routes (so it takes priority)
		vs.Spec.Http = append([]*networkingv1beta1.HTTPRoute{&newRoute}, vs.Spec.Http...)

		// Apply the updated VirtualService
		_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update VirtualService: %v", err)
		}

		log.Printf("Updated VirtualService %s with new header-based routing\n", cfg.VirtualServiceName)
		return nil
	}
*/
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

func ApplyTrafficSplit(client istioclientset.Interface, cfg config.AppConfig, v1Weight, v2Weight int32) error {
	vsClient := client.NetworkingV1alpha3().VirtualServices(cfg.Namespace)

	vs, err := vsClient.Get(context.Background(), cfg.VirtualServiceName, metaV1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VirtualService: %v", err)
	}

	log.Println("Original VirtualService fetched. Applying A/B traffic split...")

	// Define weighted routing
	splitRoute := &networkingv1beta1.HTTPRoute{
		Route: []*networkingv1beta1.HTTPRouteDestination{
			{
				Destination: &networkingv1beta1.Destination{
					Host:   cfg.Host,
					Subset: "v1",
				},
				Weight: v1Weight,
			},
			{
				Destination: &networkingv1beta1.Destination{
					Host:   cfg.Host,
					Subset: "v2",
				},
				Weight: v2Weight,
			},
		},
	}

	// Replace entire HTTP block (assumes A/B routing takes priority)
	vs.Spec.Http = []*networkingv1beta1.HTTPRoute{splitRoute}

	_, err = vsClient.Update(context.Background(), vs, metaV1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update VirtualService with traffic split: %v", err)
	}

	log.Printf("Updated VirtualService %s with v1:%d%%, v2:%d%% traffic split\n",
		cfg.VirtualServiceName, v1Weight, v2Weight)
	return nil
}
