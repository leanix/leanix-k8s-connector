package kubernetes

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cronjobs gets the list of cronjobs in a namespace
func (k *API) StatefulSets(namespace string) (*v1.StatefulSetList, error) {
	statefulSets, err := k.Client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSets, nil
}
