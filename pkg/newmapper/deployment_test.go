package newmapper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMapDeployments(t *testing.T) {

	deployment := appsv1.Deployment{
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
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "RollingUpdate",
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": "kube-dns",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"memory": resource.MustParse("170Mi"),
								},
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("70Mi"),
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}
	mapDeployment, err := DeploymentDataMapping(deployment)
	assert.NoError(t, err)
	md := mapDeployment.Data.(map[string]interface{})

	assert.Equal(t, "deployment-1-namespace:test-deployment-1 in test-cluster", md["name"])
	assert.Equal(t, "deployment-1-namespace", md["namespace"])
	assert.Equal(t, "test-cluster", md["clusterName"])
	assert.Equal(t, map[string]string{"beta.kubernetes.io/instance-type": "Standard_D8s_v3", "failure-domain.beta.kubernetes.io/region": "westeurope", "failure-domain.beta.kubernetes.io/zone": "2", "name": "nodepool-2"}, md["labels"])
	assert.Equal(t, map[string]string{
		"deployment.kubernetes.io/revision": "1"}, md["annotations"])
	assert.Equal(t, "nginx", md["image"])
	assert.Equal(t, "", md["version"])
	assert.Equal(t, "2019-01-12T08:55:20Z", md["deploymentTime"])
	assert.Equal(t, appsv1.DeploymentStrategyType("RollingUpdate"), md["updateStrategy"])
	assert.Equal(t, "nginx", md["k8sImage"])
	assert.Equal(t, int32(1), md["replicas"])
	assert.Equal(t, int32(1), md["readyReplicas"])
}
