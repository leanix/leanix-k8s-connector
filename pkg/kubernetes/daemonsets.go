package kubernetes

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSets gets the list of daemonSets in a namespace
func (k *API) DaemonSets(namespace string) (*v1.DaemonSetList, error) {
	daemonSets, err := k.Client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return daemonSets, nil
}
