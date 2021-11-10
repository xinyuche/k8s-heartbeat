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
	log.Println("Starting to find InClusterConfig")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("Failed to get InClusterConfig")
	} else {
		log.Println("Find InClusterConfig successful")
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Failed to fetch clientset by InClusterConfig")
	}

	// loop the heartbeat style test for every 5 min
	for {
		log.Println("Starting to fetch node list")
		nodelist, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Failed to get node list: %v", err)
		}
		numOfNode := len(nodelist.Items)
		log.Printf("current num of node: %v", numOfNode)
		log.Println("Cycle over, sleep for 5 mins")
		time.Sleep(5 * time.Minute)
	}
}
