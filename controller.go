package main

import (
	"context"
	"fmt"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1Informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

type Controller struct {
	clientset      kubernetes.Clientset
	podLister      corev1listers.PodLister
	cacheSynced    cache.InformerSynced
	taskQueue      workqueue.Interface
	controllerDone chan struct{}
}

func NewController(clientset kubernetes.Clientset, podInformer corev1Informers.PodInformer, controllerDone chan struct{}) *Controller {
	c := &Controller{
		clientset:      clientset,
		podLister:      podInformer.Lister(),
		cacheSynced:    podInformer.Informer().HasSynced,
		taskQueue:      workqueue.New(),
		controllerDone: controllerDone,
	}

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(addObj interface{}) {
				pod := addObj.(*core_v1.Pod)
				if pod.OwnerReferences == nil {
					c.addPod(addObj)
					fmt.Printf("Added pod %s\n", pod.Name)
				}
			},
		})

	return c
}

func (c *Controller) run() {
	fmt.Printf("Starting Controller...\n")
	fmt.Printf("Waiting for Cache to Sync...\n")
	// Wait to see if the Informer Cache has synced or initialized.
	hasSynced := cache.WaitForCacheSync(c.controllerDone, c.cacheSynced)

	if !hasSynced {
		fmt.Printf("Error Syncing Cache...\n")
	}

	fmt.Printf("Cacehe has Synced!\n")

	// Wait until the "channel" is closed
	go wait.Until(c.work, time.Second*1, c.controllerDone)

	<-c.controllerDone
}

// Get values from the work queue
func (c *Controller) work() {
	// Get a "task" from the queue.
	task, shutdown := c.taskQueue.Get()

	if shutdown {
		close(c.controllerDone)
	}

	// key is namespace/name of item
	key, _ := cache.MetaNamespaceKeyFunc(task)

	// Split key into namespace and name variables
	namespace, name, _ := cache.SplitMetaNamespaceKey(key)

	err := c.syncPod(namespace, name)

	if err != nil {
		fmt.Printf("Error syncing pod. Error: %s\n", err.Error())
		close(c.controllerDone)
	}

	c.taskQueue.Done(task)
}

func (c *Controller) syncPod(namespace, name string) error {
	// Get a pod for the namespace and name
	pod, err := c.podLister.Pods(namespace).Get(name)

	if err != nil {
		return err
	}

	// Create a ReplicaSet
	var replicas int32
	replicas = 3
	replicaSetObj := NewReplicaSet(pod, &replicas)
	replicaSet, err := c.clientset.AppsV1().ReplicaSets(namespace).Create(context.Background(), replicaSetObj, meta_v1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating ReplicaSet. Error: %s", err.Error())
		return err
	}
	fmt.Printf("-- ReplicaSet created: %s --\n", replicaSet.Name)
	return nil
}

func (c *Controller) addPod(task interface{}) {
	pod, ok := task.(*core_v1.Pod)

	if ok {
		c.taskQueue.Add(pod)
	} else {
		fmt.Printf("Error in addPod for %v", pod)
	}

}
