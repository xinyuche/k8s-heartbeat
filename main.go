package main

import (
	"context"
	"log"
	"net"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	CONN_HOST = "10.244.2.3"
	CONN_PORT = "32504"
	CONN_TYPE = "tcp"
)

type Heartbeat struct {
	clientSet *kubernetes.Clientset
}

func (hb *Heartbeat) IPFinder() error {
	pods, err := hb.clientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Unable to fetch pod list: %v", err)
		return err
	}
	for _, pod := range pods.Items {
		log.Printf("Pod Name: %v  Pod IP: %v  Pod's NodeName: %v", pod.Name, pod.Status.PodIP, pod.Spec.NodeName)
	}
	return nil
}

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

	// create Heartbeat
	hb := Heartbeat{clientSet: clientSet}
	log.Println("Fetching Pod IP list")
	iperr := hb.IPFinder()
	if iperr != nil {
		log.Println("Not able to fetch ip list")
	}

	// tcp client
	log.Print("Dialling " + CONN_HOST + ":" + CONN_PORT)
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		log.Fatalf("Dial failed to %s", CONN_PORT)
	}

	log.Print("sending heartbeat")
	conn.Write([]byte("heartbeat."))

	log.Print("waiting for response")

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	conn.Read(buf)

	log.Printf("received: %s", string(buf[:]))
	conn.Close()

	// loop the heartbeat style test for every 5 min
	for {
		log.Println("Starting to fetch node list")
		nodelist, err := hb.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Failed to get node list: %v", err)
		}
		numOfNode := len(nodelist.Items)
		log.Printf("current num of node: %v", numOfNode)
		log.Println("Cycle over, sleep for 5 mins")
		time.Sleep(5 * time.Minute)
	}
}
