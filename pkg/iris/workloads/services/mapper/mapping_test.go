package mapper

import (
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"
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
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-cronjob-1",
				Namespace:         "cronjob-1-namespace",
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
			Spec: batchv1.CronJobSpec{
				Schedule: "0 0 * * *",
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "service-2"},
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
				},
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-statefulset-1",
				Namespace:         "statefulset-1-namespace",
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
			Spec: appsv1.StatefulSetSpec{
				Replicas:    pointer.Int32Ptr(1),
				ServiceName: "test-service-1",
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
			Status: appsv1.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-daemonset-1",
				Namespace:         "daemonset-1-namespace",
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
			Spec: appsv1.DaemonSetSpec{
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
		},
	}
	mockApi := kubernetes.API{
		Client: fake.NewSimpleClientset(append(dummyDeployments, dummyServices...)...),
	}
	mapper := NewMapper(&mockApi, "testCluster", "testWorkspace", "testRunId")
	results, err := mapper.MapWorkloads("testCluster")

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, 5, len(results))
	assert.Equal(t, "test-deployment-1", results[0].WorkloadName)
	assert.Equal(t, "service-2", results[0].ServiceName)
	assert.Equal(t, "deployment", results[1].WorkloadType)

	// test mapping cronjob
	assert.Equal(t, "test-cronjob-1", results[2].WorkloadName)
	assert.Equal(t, "service-2", results[2].ServiceName)
	assert.Equal(t, "cronjob", results[2].WorkloadType)
	assert.Equal(t, "0 0 * * *", results[2].WorkloadProperties.Schedule)

	// test mapping statefulset
	assert.Equal(t, "test-statefulset-1", results[3].WorkloadName)
	assert.Equal(t, "service-1", results[3].ServiceName)
	assert.Equal(t, "statefulSet", results[3].WorkloadType)
	assert.Equal(t, string("\x01"), results[3].WorkloadProperties.Replicas)

	// test mapping daemonset
	assert.Equal(t, "test-daemonset-1", results[4].WorkloadName)
	assert.Equal(t, "service-1", results[4].ServiceName)
	assert.Equal(t, "daemonSet", results[4].WorkloadType)
	assert.Equal(t, "50", results[4].Containers.K8sLimits.Memory)

}
