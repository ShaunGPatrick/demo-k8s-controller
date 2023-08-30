package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	// Path to kubeconfig saved in an env variable.
	// e.g. ~/.kube/config
	// kubeconfig holds the current state of all resources in your cluster.
	kubeConfig := os.Getenv("KUBE_CONFIG")

	// Get kubeConfig and clientSet.
	config, _ := clientcmd.BuildConfigFromFlags("", kubeConfig)
	clientSet, _ := kubernetes.NewForConfig(config)

	// Getting a specific deployment from a namespace.
	getDeployment(clientSet, "default", "coffee")

	// Listing deployments in a namespace.
	listDeployments(clientSet, "default")

	// Creating informers for deployments.
	sharedInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*30)
	getDeploymentFromLister(sharedInformerFactory)
	deploymentEventHandler(sharedInformerFactory)
}

func getDeployment(clientset *kubernetes.Clientset, namespace, deploymentName string) {

	deployment, _ := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, v1.GetOptions{})

	fmt.Printf("Found deployment %s\n", deployment.ObjectMeta.Name)

}

func listDeployments(clientset *kubernetes.Clientset, namespace string) {

	fmt.Printf("Getting deployments in namespace %s\n", namespace)

	deployments, _ := clientset.AppsV1().Deployments(namespace).List(context.TODO(), v1.ListOptions{})

	for _, deployment := range deployments.Items {
		fmt.Printf("Deployment name: %s\n", deployment.ObjectMeta.Name)
		fmt.Printf("Deployment replicas: %d\n", deployment.Spec.Replicas)
	}
}

func getDeploymentFromLister(sharedInformerFactory informers.SharedInformerFactory) {
	deploymentInformer := sharedInformerFactory.Apps().V1().Deployments()
	deployments, _ := deploymentInformer.Lister().Deployments("default").Get("coffee")

	fmt.Printf("Deployments %v", deployments)
}

func deploymentEventHandler(sharedInformerFactory informers.SharedInformerFactory) {

	deploymentInformer := sharedInformerFactory.Apps().V1().Deployments()

	fmt.Println("Waiting for deployment events...")

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(addObj interface{}) {
			deployment := addObj.(*apps_v1.Deployment)
			fmt.Printf("Deployment added %s \n", deployment.ObjectMeta.Name)
		},
		DeleteFunc: func(delObj interface{}) {
			deployment := delObj.(*apps_v1.Deployment)
			fmt.Printf("Deployment deleted %s \n", deployment.ObjectMeta.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				newDeployment := newObj.(*apps_v1.Deployment)
				fmt.Printf("Deployment updated %v \n", newDeployment.ObjectMeta.Name)
			} else {
				fmt.Println("UpdateFunc: No difference in deployments")
			}
		},
	})
}
