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
	namespace := "autoreplica"

	// 1. Get kubeConfig and clientSet.
	config, _ := clientcmd.BuildConfigFromFlags("", kubeConfig)
	clientSet, _ := kubernetes.NewForConfig(config)

	// 2. Initialize informer factory in a specific namespace.
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientSet,
		time.Second*30,
		informers.WithNamespace(namespace))

	// 3. Make empty channel
	channel := make(chan struct{})

	// 4. Initialize a new controller
	controller := NewController(*clientSet, sharedInformerFactory.Core().V1().Pods(), channel)

	// 5. Start informer
	sharedInformerFactory.Start(channel)

	// 6. Run controller
	controller.run()
}
