package kube

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func MustClient() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Println("Falling back to in-cluster config")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error loading kube config: %v", err)
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating k8s client: %v", err)
	}
	return clientset
}
