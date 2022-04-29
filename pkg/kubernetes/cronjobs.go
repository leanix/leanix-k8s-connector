package kubernetes

import (
	"context"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cronjobs gets the list of cronjobs in a namespace
func (k *API) CronJobs(namespace string) (*v1.CronJobList, error) {
	cronJobs, err := k.Client.BatchV1().CronJobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return cronJobs, nil
}
