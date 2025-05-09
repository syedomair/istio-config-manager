package kube

import (
	"log"

	istioClientset "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Clients struct {
	Kube  *kubernetes.Clientset
	Istio *istioClientset.Clientset
}

func MustClients() *Clients {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Println("Falling back to in-cluster config")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error loading kube config: %v", err)
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating k8s client: %v", err)
	}

	istioClient, err := istioClientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Istio client: %v", err)
	}

	return &Clients{
		Kube:  kubeClient,
		Istio: istioClient,
	}
}
