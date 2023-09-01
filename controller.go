package main

import (
	"fmt"
	apps_v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

type Controller struct {
	clientset             kubernetes.Clientset
	appLister             appsv1listers.DeploymentLister
	deploymentCacheSynced cache.InformerSynced
	taskQueue             workqueue.RateLimitingInterface
}

func NewController(clientset kubernetes.Clientset, deploymentInformer v1.DeploymentInformer) *Controller {
	c := &Controller{
		clientset:             clientset,
		appLister:             deploymentInformer.Lister(),
		deploymentCacheSynced: deploymentInformer.Informer().HasSynced,
		taskQueue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	deploymentInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(addObj interface{}) {
				deployment := addObj.(*apps_v1.Deployment)
				fmt.Printf("Deployment was added: %s\n", deployment.ObjectMeta.Name)
			},
			DeleteFunc: func(delObj interface{}) {
				deployment := delObj.(*apps_v1.Deployment)
				fmt.Printf("Deployment was deleted: %s\n", deployment.ObjectMeta.Name)
			},
		})

	return c
}

func (c *Controller) run(channel <-chan struct{}) {
	fmt.Printf("Starting Controller...\n")
	fmt.Printf("Waiting for Cache to Sync...\n")
	// Wait to see if the Informer Cache has synced or initialized.
	hasSynced := cache.WaitForCacheSync(channel, c.deploymentCacheSynced)

	if !hasSynced {
		fmt.Printf("Error Syncing Cache...\n")
	}

	// Wait until the "channel" is closed
	go wait.Until(c.work, time.Second*1, channel)

	<-channel
}

func (c *Controller) work() {

}
