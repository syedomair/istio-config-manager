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
