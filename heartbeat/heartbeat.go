package heartbeat

import (
	"k8s.io/client-go/kubernetes"
)

type heartbeat struct {
	clientSet *kubernetes.Clientset
}

func (hb *heartbeat) ScaleDownNode() {
	// 1. have current num of nodes
	// 2. have the test app num of pods in the cluster
	// 3. halt the node and see if the num of pods can be back to the desired state
}
