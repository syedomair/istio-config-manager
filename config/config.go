package config

import (
	"log"
	"os"
)

type AppConfig struct {
	VirtualServiceName string
	Namespace          string
	Host               string
	Destination        string
}

func LoadConfig() AppConfig {
	return AppConfig{
		VirtualServiceName: getEnv("VIRTUAL_SERVICE_NAME", "my-service"),
		Namespace:          getEnv("NAMESPACE", "default"),
		Host:               getEnv("HOST", "my-service.local"),
		Destination:        getEnv("DESTINATION", "v1"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	log.Printf("Env %s not set. Using default: %s", key, fallback)
	return fallback
}
