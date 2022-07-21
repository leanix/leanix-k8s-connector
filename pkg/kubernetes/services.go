package kubernetes

import (
	"context"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Services gets the list of services in a namespace
func (k *API) Services(namespace string) (*corev1.ServiceList, error) {
	services, err := k.Client.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return services, nil
}
