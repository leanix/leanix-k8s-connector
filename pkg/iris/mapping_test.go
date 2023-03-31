package iris

import (
	"testing"
	"time"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMapDeployments(t *testing.T) {
	// create a dummy nodes
	dummyDeployments := appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-deployment-2",
					Namespace:         "deployment-1-namespace",
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
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":  "app2",
							"test": "false",
						},
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "testImage",
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
											corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
										}}},
							},
						},
					},
				},
				Status: appsv1.DeploymentStatus{
					Replicas:      1,
					ReadyReplicas: 1,
				},
			},
		},
	}
	// create dummy services
	dummyServices := corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-1",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "app1",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-2",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app":  "app2",
						"test": "false",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
		},
	}
	mapper := NewMapper(&kubernetes.API{}, "cluster-test", "workspace-test", make([]string, 0), "test-runid")
	resultEcst, err := mapper.MapDeploymentsEcst(&dummyDeployments, &dummyServices)

	assert.NoError(t, err)
	assert.NotEmpty(t, resultEcst)
	assert.Equal(t, "test-deployment-2", resultEcst[0].DeploymentName)
	assert.Equal(t, "test-service-2", resultEcst[0].ServiceName)
	assert.Equal(t, "100", resultEcst[0].DeploymentProperties.K8sRequests.Cpu)
	assert.Equal(t, "50", resultEcst[0].DeploymentProperties.K8sRequests.Memory)
	assert.Equal(t, "", resultEcst[0].DeploymentProperties.K8sLimits.Cpu)
	assert.Equal(t, "", resultEcst[0].DeploymentProperties.K8sLimits.Memory)
	assert.Equal(t, "1", resultEcst[0].DeploymentProperties.Replicas)
	assert.Equal(t, "testImage", resultEcst[0].Image)
}

func TestMapDeployments_NoService(t *testing.T) {
	// create a dummy nodes
	dummyDeployments := appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-deployment-2",
					Namespace:         "deployment-1-namespace",
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
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":  "app2",
							"test": "false",
						},
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "testImage",
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
											corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
										}}},
							},
						},
					},
				},
				Status: appsv1.DeploymentStatus{
					Replicas:      1,
					ReadyReplicas: 1,
				},
			},
		},
	}
	// create dummy services
	dummyServices := corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-1",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "app1",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
		},
	}
	mapper := NewMapper(&kubernetes.API{}, "cluster-test", "workspace-test", make([]string, 0), "test-runid")
	resultEcst, err := mapper.MapDeploymentsEcst(&dummyDeployments, &dummyServices)

	assert.NoError(t, err)
	assert.NotEmpty(t, resultEcst)

	assert.Equal(t, "test-deployment-2", resultEcst[0].DeploymentName)
	assert.Empty(t, resultEcst[0].ServiceName)
	assert.Equal(t, "100", resultEcst[0].DeploymentProperties.K8sRequests.Cpu)
	assert.Equal(t, "50", resultEcst[0].DeploymentProperties.K8sRequests.Memory)
	assert.Equal(t, "", resultEcst[0].DeploymentProperties.K8sLimits.Cpu)
	assert.Equal(t, "", resultEcst[0].DeploymentProperties.K8sLimits.Memory)
	assert.Equal(t, "1", resultEcst[0].DeploymentProperties.Replicas)
	assert.Equal(t, "testImage", resultEcst[0].Image)
}

func TestGetCluster(t *testing.T) {
	// create a dummy nodes
	dummyNodes := corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-node-1",
					Namespace:         "test-namepsace",
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
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						KubeletVersion: "abc",
						OSImage:        "123",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-node-2",
					Namespace:         "test-namepsace",
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
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						KubeletVersion: "def",
						OSImage:        "456",
					},
				},
			},
		},
	}

	mapper := NewMapper(&kubernetes.API{}, "test-cluster", "workspace-test", make([]string, 0), "test-runid")
	result, err := mapper.MapCluster("test-cluster", &dummyNodes)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.Equal(t, "test-cluster", result.name)
	assert.Equal(t, 2, result.nodesCount)
	assert.Contains(t, result.k8sVersion, "def")
	assert.Contains(t, result.k8sVersion, "abc")
	assert.Contains(t, result.osImage, "123")
	assert.Contains(t, result.osImage, "456")
}

func TestResolveServiceForDeployment_Success(t *testing.T) {
	dummyServices := corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-1",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "app1",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-2",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app":  "app2",
						"test": "false",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
		},
	}

	dummyDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-deployment-2",
			Namespace:         "deployment-1-namespace",
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
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "app1",
					"test": "false",
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
								corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
							}}},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}
	result := ResolveK8sServiceForK8sDeployment(&dummyServices, dummyDeployment)
	assert.NotEmpty(t, result)
	assert.Equal(t, "test-service-1", result)
}

func TestResolveServiceForDeployment_NoCommonLabels(t *testing.T) {
	dummyServices := corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-1",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "app1",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-2",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "app2",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
		},
	}

	dummyDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-deployment-2",
			Namespace:         "deployment-1-namespace",
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
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"production": "ready",
					"test":       "false",
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
								corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
							}}},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}
	result := ResolveK8sServiceForK8sDeployment(&dummyServices, dummyDeployment)
	assert.Equal(t, "", result)
}

func TestResolveServiceForDeployment_DifferentLabelValues(t *testing.T) {
	dummyServices := corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-1",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "app1",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-service-2",
					Namespace:         "deployment-1-namespace",
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
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app":  "app2",
						"test": "false",
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{},
					Conditions:   nil,
				},
			},
		},
	}

	dummyDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-deployment-2",
			Namespace:         "deployment-1-namespace",
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
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "app3",
					"test": "false",
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(int64(100), "Mi"),
								corev1.ResourceMemory: *resource.NewQuantity(int64(50), "m"),
							}}},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}
	result := ResolveK8sServiceForK8sDeployment(&dummyServices, dummyDeployment)
	assert.Equal(t, "", result)
}
