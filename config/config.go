package config

import (
	"log"
	"os"
)

type AppConfig struct {
	VirtualServiceName  string
	Namespace           string
	Host                string
	Destination         string
	SubsetName          string
	SkipVirtualService  bool
	SkipDestinationRule bool
	HeaderName          string
	HeaderValue         string
	TargetPrefix        string
	TargetRewrite       string
	TrafficWeight1      string
	TrafficWeight2      string
	FaultDelayMillis    string
	TimeoutDuration     string
	RetryAttempt        string
	RetryPerTryTimeout  string
	RetryOn             string
}

func LoadConfig() AppConfig {
	return AppConfig{
		VirtualServiceName:  getEnv("VIRTUAL_SERVICE_NAME", "user-service"),
		Namespace:           getEnv("NAMESPACE", "default"),
		Host:                getEnv("HOST", "user-service"),
		Destination:         getEnv("DESTINATION", "v1"),
		SubsetName:          getEnv("SUBSET_NAME", "v2"),
		SkipVirtualService:  getEnvBool("SKIP_VIRTUAL_SERVICE", true),
		SkipDestinationRule: getEnvBool("SKIP_DESTINATION_RULE", true),
		HeaderName:          getEnv("HEADER_NAME", "x-user-type"),
		HeaderValue:         getEnv("HEADER_VALUE", "beta"),
		TargetPrefix:        getEnv("TARGET_PREFIX", "/user/users"),
		TargetRewrite:       getEnv("TARGET_REWRITE", "/v2/users"),
		TrafficWeight1:      getEnv("WEIGHT_V1", "90"),
		TrafficWeight2:      getEnv("WEIGHT_V2", "10"),
		FaultDelayMillis:    getEnv("FAULT_DELAY_MILLIS", "5000"),
		TimeoutDuration:     getEnv("TIMEOUT_DURATION", "2"),
		RetryAttempt:        getEnv("RETRY_ATTEMPT", "3"),
		RetryPerTryTimeout:  getEnv("RETRY_PER_TRY_TIMEOUT", "1"),
		RetryOn:             getEnv("RETRY_ON", "gateway-error,connect-failure,refused-stream,5xx"),
	}
}

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val == "true" || val == "1"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	log.Printf("Env %s not set. Using default: %s", key, fallback)
	return fallback
}
