package main

import (
	"context"
	"fmt"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
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

	// - print out the namespaces on the cluster
	// Note on `context.Context` in go: https://pauldigian.com/golang-and-context-an-explanation
	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("This cluster has the following namespaces")
	for _, namespace := range namespaces.Items {
		fmt.Println(namespace.Name)
	}

	// - create a new namespace
	newNamespace, err := clientSet.CoreV1().Namespaces().Create(
		context.TODO(),
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "new-test-namespace"}},
		metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	// - create a pod in that namespace that runs a simple hello-world container
	// Define the pod
	podDefinition := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-word",
			Namespace: newNamespace.Name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "hello-world",
					Image:   "alpine:latest",
					Command: []string{"bash", "-c", "while true; do echo hello world!; sleep 1; done"},
				},
			},
		},
	}
	// Create it
	newPod, err := clientSet.CoreV1().Pods(newNamespace.Name).Create(
		context.TODO(),
		&podDefinition,
		metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	// - print out pod names and the namespace they are in for any pods that have a label of ‘k8s-app=kube-dns’ or a similar label is ok as well
	labelSelector := "k8s-app=kube-dns"
	pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		panic(err)
	}

	fmt.Println("This cluster has the following pods that match labelSelector:", labelSelector)
	for _, pod := range pods.Items {
		fmt.Println("Namespace:", pod.Namespace, "Pod:", pod.Name)
	}

	// - delete the hello-world pod created from above
	err = clientSet.CoreV1().Pods(newPod.Namespace).Delete(context.TODO(), newPod.Name, metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}

	// (might as well delete the namespace we created as well)
	err = clientSet.CoreV1().Namespaces().Delete(context.TODO(), newNamespace.Name, metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}

	// - extra credit - show how an client-go informer works
}
