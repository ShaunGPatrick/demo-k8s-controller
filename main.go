package main

import (
	"os"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	// Path to kubeconfig saved in an env variable.
	// e.g. ~/.kube/config
	// kubeconfig holds the current state of all resources in your cluster.
	kubeConfig := os.Getenv("KUBE_CONFIG")

	// 1. Get kubeConfig and clientSet.
	config, _ := clientcmd.BuildConfigFromFlags("", kubeConfig)
	clientSet, _ := kubernetes.NewForConfig(config)

	// 2. Initialize informer factory.
	sharedInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*30)

	// 3. Empty channel
	channel := make(chan struct{})

	// 4. Initialize a new controller
	controller := NewController(*clientSet, sharedInformerFactory.Apps().V1().Deployments())

	// 5. Start informer
	sharedInformerFactory.Start(channel)

	// 6. Run controller
	controller.run(channel)
}
