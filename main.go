package main

import (
	"context"
	"log"
	"net"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	HB_PORT   = "32504"
	CONN_TYPE = "tcp"
)

type Heartbeat struct {
	clientSet *kubernetes.Clientset
}

func (hb *Heartbeat) PodsFinder() ([]v1.Pod, error) {
	pods, err := hb.clientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Unable to fetch pod list: %v", err)
		return nil, err
	}
	for _, pod := range pods.Items {
		log.Printf("Pod Name: %v  Pod IP: %v  Pod's NodeName: %v", pod.Name, pod.Status.PodIP, pod.Spec.NodeName)
	}
	return pods.Items, nil
}

func (hb *Heartbeat) HeartbeatSender(podName string, podIP string, nodeName string) (error, bool) {
	log.Printf("Sending heartbeat to node %v. podIP: %v", nodeName, podIP)
	log.Print("Dialling " + podIP + ":" + HB_PORT)
	hbconn, err := net.Dial(CONN_TYPE, podIP+":"+HB_PORT)
	if err != nil {
		log.Printf("Dial failed to port %s", HB_PORT)
		return err, false
	}
	log.Println("Writing heartbeat to conn")
	hbconn.Write([]byte("heartbeat"))
	log.Println("Waiting for response")
	// Make a buffer to hold incoming data.
	hbbuf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	hbconn.Read(hbbuf)
	log.Printf("Received: %s", string(hbbuf[:]))
	if string(hbbuf[:]) == "heartbeat received" {
		hbconn.Close()
		log.Println(nodeName + " is running")
		return nil, true
	} else {
		hbconn.Write([]byte("restart"))
		log.Println(nodeName + " is requested for restart")
		hbconn.Close()
		return nil, false
	}
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

	// loop the heartbeat style test for every 5 min
	for {
		log.Println("Starting to fetch node list")
		nodelist, err := hb.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Failed to get node list: %v", err)
		}
		numOfNode := len(nodelist.Items)
		log.Printf("current num of node: %v", numOfNode)

		log.Println("Fetching Pod IP list")
		podList, podListErr := hb.PodsFinder()
		if podListErr != nil {
			log.Println("Not able to fetch ip list")
		}

		for _, pod := range podList {
			var podName = pod.Name
			var podIP = pod.Status.PodIP
			var nodeName = pod.Spec.NodeName
			log.Printf("Pod Name: %v  Pod IP: %v  Pod's NodeName: %v", pod.Name, pod.Status.PodIP, pod.Spec.NodeName)
			if nodeName != "minikube" {
				err, status := hb.HeartbeatSender(podName, podIP, nodeName)
				if err != nil {
					log.Printf("Send Heartbeat failed: %v", err)
				}
				if status {
					log.Println("Node: " + nodeName + " is active ")
				} else {
					log.Println("Node: " + nodeName + " is not active ")
				}
			}
		}
		log.Println("Cycle over, sleep for 5 mins")
		time.Sleep(5 * time.Minute)
	}
}
