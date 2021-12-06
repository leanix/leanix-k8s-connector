package kubernetes

import (
	"context"
	"github.com/leanix/leanix-k8s-connector/pkg/set"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Nodes gets the list of worker nodes (kubelets)
func (k *API) Nodes() (*corev1.NodeList, error) {
	nodes, err := k.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// NodesByName returns a list of nodes for the given node names
func (k *API) NodesByName(nodeNames *set.String) (*[]corev1.Node, error) {
	nodes, err := k.Nodes()
	if err != nil {
		return nil, err
	}
	matches := make([]corev1.Node, 0)
	for _, n := range nodes.Items {
		if nodeNames.Contains(n.Name) {
			matches = append(matches, n)
		}
	}
	return &matches, nil
}
