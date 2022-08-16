package kubernetes

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicateSets gets the list of replicaSets in a namespace
func (k *API) ReplicaSets(namespace string) (*v1.ReplicaSetList, error) {
	replicaSets, err := k.Client.AppsV1().ReplicaSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return replicaSets, nil
}
