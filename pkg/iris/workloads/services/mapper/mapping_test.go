package mapper

import (
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func Test_MapWorkloads_success(t *testing.T) {
	dummyServices := []runtime.Object{
		&corev1.Service{
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "service-1"},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "service-1",
				Labels: map[string]string{
					"name": "service-1",
					"failure-domain.beta.kubernetes.io/region": "westeurope",
					"failure-domain.beta.kubernetes.io/zone":   "1",
					"beta.kubernetes.io/instance-type":         "Standard_D2s_v3",
				},
			},
		},
		&corev1.Service{
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "service-2"},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "service-2",
				Labels: map[string]string{
					"name": "nodepool-2",
					"failure-domain.beta.kubernetes.io/region": "westeurope",
					"failure-domain.beta.kubernetes.io/zone":   "2",
					"beta.kubernetes.io/instance-type":         "Standard_D8s_v3",
				},
			},
		},
	}

	dummyDeployments := []runtime.Object{
		&appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
										corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
										corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
									},
								},
							},
						},
					},
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "service-2"},
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-deployment-1",
				Namespace:         "deployment-1-namespace",
				CreationTimestamp: metav1.Date(2019, 01, 12, 8, 55, 20, 0, time.UTC),
				Labels: map[string]string{
					"name": "service-2",
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
				CreationTimestamp: metav1.Date(2019, 01, 12, 8, 55, 20, 0, time.UTC),
				Labels: map[string]string{
					"name": "service-1",
					"failure-domain.beta.kubernetes.io/region": "westeurope",
					"failure-domain.beta.kubernetes.io/zone":   "2",
					"beta.kubernetes.io/instance-type":         "Standard_D8s_v3",
				},
				Annotations: map[string]string{
					"deployment.kubernetes.io/revision": "1",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "service-1"},
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
										corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
										corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
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
		},
	}
	mockApi := kubernetes.API{
		Client: fake.NewSimpleClientset(append(dummyDeployments, dummyServices...)...),
	}
	mapper := NewMapper(&mockApi, "testCluster", "testWorkspace", "testRunId")
	results, err := mapper.MapWorkloads("testCluster")

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}
