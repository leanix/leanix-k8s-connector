package iris

import (
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	fakev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var nodeList = corev1.NodeList{
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

var namespaceList = corev1.NamespaceList{
	Items: []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"},
		},
	},
}

var deploymentList = appsv1.DeploymentList{
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
		},
	},
}

func TestScan(t *testing.T) {
	setup()
	apiMock := func(config *rest.Config) (*kubernetes.API, error) {
		client := fake.NewSimpleClientset()
		client.CoreV1().(*fakev1.FakeCoreV1).
			PrependReactor("list", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &nodeList, nil
			})
		client.CoreV1().(*fakev1.FakeCoreV1).
			PrependReactor("list", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &namespaceList, nil
			})
		client.CoreV1().(*fakev1.FakeCoreV1).
			PrependReactor("list", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &deploymentList, nil
			})
		k := kubernetes.API{
			Client: client,
		}
		return &k, nil
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		rw.Write([]byte(`{"cluster":"abc"}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	scanner := NewScanner("test-kind", server.URL, "test-run-id", "test-token", "test-workspace")
	err := scanner.Scan(apiMock, nil, "test-token")

	assert.NoError(t, err)

}
