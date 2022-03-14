package kubernetes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeployments(t *testing.T) {
	// create a dummy nodes
	dummyDeployments := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-deployment-1",
				Namespace:         "deployment-1-namespace",
				ClusterName:       "test-cluster",
				CreationTimestamp: metav1.Date(2019, 01, 12, 8, 55, 20, 0, time.UTC),
				Labels: map[string]string{
					"name": "nodepool-2",
					"failure-domain.beta.kubernetes.io/region": "westeurope",
					"failure-domain.beta.kubernetes.io/zone":   "2",
					"beta.kubernetes.io/instance-type":         "Standard_D8s_v3",
				},
				Annotations: map[string]string{
					"deployment.kubernetes.io/revision": "1",
				},
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-deployment-2",
				Namespace:         "deployment-1-namespace",
				ClusterName:       "test-cluster",
				CreationTimestamp: metav1.Date(2019, 01, 12, 8, 55, 20, 0, time.UTC),
				Labels: map[string]string{
					"name": "nodepool-2",
					"failure-domain.beta.kubernetes.io/region": "westeurope",
					"failure-domain.beta.kubernetes.io/zone":   "2",
					"beta.kubernetes.io/instance-type":         "Standard_D8s_v3",
				},
				Annotations: map[string]string{
					"deployment.kubernetes.io/revision": "1",
				},
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		},
	}
	k := API{
		Client: fake.NewSimpleClientset(dummyDeployments...),
	}

	deployments, err := k.Deployments("deployment-1-namespace")
	if err != nil {
		t.Error(err)
	}

	assert.Len(t, deployments.Items, 2)
}
