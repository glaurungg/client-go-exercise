package main

import (
	"fmt"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	// - connect to the k8s cluster

	// minkube sets up the kubeconfig in the default location, but we can assume in this code we have a
	// valid kubeconfig in ~/.kube/config with a currently set current-context
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	// NOTE: if we were executing this from inside the cluster, we could get our kubeconfig from `rest.InClusterConfig()`,
	// which looks at the default mounted /var/run/secrets/kubernetes.io/serviceaccount path
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	endChan := make(chan struct{})

	// Use a factory so all watchers in the process space can share the cached store, further reducing load on api server
	informerFactory := informers.NewSharedInformerFactory(clientSet, 0)

	// grab / instantiate a shared namespace informer
	namespaceInformer := informerFactory.Core().V1().Namespaces()

	// ... and add a callback that will get fired when a new namespace is seen
	namespaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("New namespace added:", obj.(*v1.Namespace).ObjectMeta.Name)
		},
		// note we are explicitly not adding update or delete handlers, but we would definitely want to in the future
	})

	fmt.Println("Started informer")
	namespaceInformer.Informer().Run(endChan)

	// Just chill for 5 minutes
	end := time.Now().Add(60 * 5 * time.Second)
	for time.Now().Before(end) {
		time.Sleep(time.Second)
	}

	close(endChan)
	fmt.Println("Stopped informer, exiting")
}
