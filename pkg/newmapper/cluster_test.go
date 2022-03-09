package newmapper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAggregateNodes(t *testing.T) {
	nodes := &corev1.NodeList{
		Items: []corev1.Node{
			corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						Architecture:            "amd64",
						ContainerRuntimeVersion: "docker://3.0.1",
						KernelVersion:           "4.15.0-1035-azure",
						KubeletVersion:          "v1.11.5",
						OperatingSystem:         "linux",
						OSImage:                 "Ubuntu 16.04.5 LTS",
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Date(2019, 01, 18, 8, 55, 20, 0, time.UTC),
					Name:              "nodepool-1",
					Labels: map[string]string{
						"name": "nodepool-1",
						"failure-domain.beta.kubernetes.io/region": "westeurope",
						"failure-domain.beta.kubernetes.io/zone":   "1",
						"beta.kubernetes.io/instance-type":         "Standard_D2s_v3",
					},
				},
			},
			corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						Architecture:            "amd64",
						ContainerRuntimeVersion: "docker://3.0.1",
						KernelVersion:           "4.15.0-1035-azure",
						KubeletVersion:          "v1.11.5",
						OperatingSystem:         "linux",
						OSImage:                 "Ubuntu 16.04.5 LTS",
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Date(2019, 01, 12, 8, 55, 20, 0, time.UTC),
					Name:              "nodepool-2",
					Labels: map[string]string{
						"name": "nodepool-2",
						"failure-domain.beta.kubernetes.io/region": "westeurope",
						"failure-domain.beta.kubernetes.io/zone":   "2",
						"beta.kubernetes.io/instance-type":         "Standard_D8s_v3",
					},
				},
			},
		},
	}
	nodeAggregate, err := AggregrateNodes(nodes)
	assert.NoError(t, err)

	assert.Equal(t, 2, nodeAggregate["nodesCount"])
	assert.ElementsMatch(t, []string{"v1.11.5"}, nodeAggregate["k8sVersion"])
	assert.Equal(t, "Ubuntu 16.04.5 LTS", nodeAggregate["osImage"])
}
