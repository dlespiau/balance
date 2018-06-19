package balance

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func makeInClusterClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
