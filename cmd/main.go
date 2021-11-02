package main

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("Failed to get InClusterConfig")
	}
	clientSet, err := kubernetes.NewForConfig(config)

	// loop the heartbeat style test for every 5 min
	for {
		nodelist, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Println("Failed to get node list")
		}
		numOfNode := len(nodelist.Items)
		log.Printf("current num of node: %v", numOfNode)
		time.Sleep(5 * time.Minute)
	}
}
