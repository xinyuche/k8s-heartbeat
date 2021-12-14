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

type ClusterSet struct {
	clientSet *kubernetes.Clientset
}

type PodHeartbeat struct {
	PodName  string
	PodIp    string
	NodeName string
	HBConn   net.Conn
	CtlConn  net.Conn
}

func (cs *ClusterSet) PodsFinder() ([]v1.Pod, error) {
	pods, err := cs.clientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Unable to fetch pod list: %v", err)
		return nil, err
	}
	return pods.Items, nil
}

func (phb *PodHeartbeat) HeartbeatSender() (bool, error) {
	log.Printf("Sending heartbeat to node %v. podIP: %v", phb.NodeName, phb.PodIp)
	// log.Print("Dialling " + phb.PodIp + ":" + HB_PORT)
	hbconn, err := net.Dial(CONN_TYPE, phb.PodIp+":"+HB_PORT)
	phb.HBConn = hbconn
	if err != nil {
		log.Printf("Dial failed to port %s, ip: %v", HB_PORT, phb.PodIp)
		return false, err
	}
	// log.Println("Writing heartbeat to conn")
	phb.HBConn.Write([]byte("heartbeat"))
	// log.Println("Waiting for response")
	// Make a buffer to hold incoming data.
	hbbuf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	phb.HBConn.Read(hbbuf)
	if string(hbbuf[:18]) == "heartbeat received" {
		log.Println("In heartbeat received condition")
		log.Printf("Received: %s", string(hbbuf[:18]))
		// log.Println(phb.NodeName + " is running")
		phb.HBConn.Close()
		return true, nil
	} else { // failed condition can't close the conn immediatly  if string(hbbuf[:6]) == "failed"
		log.Println("In failed condition")
		log.Printf("Received: %s", string(hbbuf[:6]))
		phb.HBConn.Close()
		ctlConn, err := net.Dial(CONN_TYPE, phb.PodIp+":"+HB_PORT)
		if err != nil {
			log.Printf("Failed to send ctl signal. Dial failed to port %s, ip: %v", HB_PORT, phb.PodIp)
			return false, err
		}
		phb.CtlConn = ctlConn
		phb.CtlConn.Write([]byte("restart"))
		log.Println(phb.NodeName + " is requesting for restart")
		phb.CtlConn.Close()
		return false, nil
	}
}

func main() {
	log.Println("version 3")
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

	// create ClientSet
	cs := ClusterSet{clientSet: clientSet}

	// loop the heartbeat style test for every 5 min
	for {
		// log.Println("Starting to fetch node list")
		// nodelist, err := cs.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		// if err != nil {
		// 	log.Printf("Failed to get node list: %v", err)
		// }
		// numOfNode := len(nodelist.Items)
		// log.Printf("current num of node: %v", numOfNode)
		log.Println("Heartbeat cycle start. Fetching Pods' IP list")
		podList, podListErr := cs.PodsFinder()
		if podListErr != nil {
			log.Println("Not able to fetch ip list")
		}

		for _, pod := range podList {
			var podName = pod.Name
			var podIP = pod.Status.PodIP
			var nodeName = pod.Spec.NodeName
			log.Printf("Pod Name: %v  Pod IP: %v  Pod's NodeName: %v", pod.Name, pod.Status.PodIP, pod.Spec.NodeName)
			phb := PodHeartbeat{PodName: podName, PodIp: podIP, NodeName: nodeName}
			if phb.NodeName != "minikube" {
				status, err := phb.HeartbeatSender()
				if err != nil {
					log.Printf("Send Heartbeat failed: %v", err)
					log.Println("Node: " + phb.NodeName + " is not active ")
				}
				if status {
					log.Println("Node: " + phb.NodeName + " is active ")
				} else {
					log.Println("Node: " + phb.NodeName + " is not active ")
				}
			}
		}
		log.Println("Cycle over, sleep for 2 mins")
		time.Sleep(1 * time.Minute)
	}
}
