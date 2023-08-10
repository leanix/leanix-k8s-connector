package kubernetes

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Namespaces gets the list of blacklisted namespaces
func (k *API) Namespaces(blacklistedNamespaces []string) (*v1.NamespaceList, error) {
	namespaces, err := k.Client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{FieldSelector: NamespaceBlacklistFieldSelector(blacklistedNamespaces)})
	if err != nil {
		return nil, err
	}
	return namespaces, err
}
