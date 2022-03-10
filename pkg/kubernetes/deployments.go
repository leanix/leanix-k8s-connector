package kubernetes

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployments gets the list of deployments in a namespace
func (k *API) Deployments(namespace string) (*v1.DeploymentList, error) {
	deployments, err := k.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deployments, nil
}
