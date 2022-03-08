package newmapper

import (
	"strings"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	"github.com/leanix/leanix-k8s-connector/pkg/set"
	corev1 "k8s.io/api/core/v1"
)

// MapNodes maps a list of nodes and a given cluster name into a KubernetesObject.
// In the process it aggregates the information from muliple nodes into one cluster object.
func GetCluster(clusterName string, kubernetesAPI *kubernetes.API) (*mapper.KubernetesObject, error) {

	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		return nil, err
	}
	nodeAggregate, err := AggregrateNodes(nodes)
	if err != nil {
		return nil, err
	}
	nodeAggregate["clusterName"] = clusterName
	return &mapper.KubernetesObject{
		ID:   clusterName,
		Type: "Cluster",
		Data: nodeAggregate,
	}, nil
}

func AggregrateNodes(nodes *corev1.NodeList) (map[string]interface{}, error) {
	nodeAggregate := make(map[string]interface{})
	items := nodes.Items
	if len(items) == 0 {
		return nodeAggregate, nil
	}
	k8sVersion := set.NewStringSet()
	osImage := set.NewStringSet()

	for _, n := range items {
		k8sVersion.Add(n.Status.NodeInfo.KubeletVersion)
		osImage.Add(n.Status.NodeInfo.OSImage)

	}
	nodeAggregate["k8sVersion"] = k8sVersion.Items()
	nodeAggregate["nodesCount"] = len(items)
	nodeAggregate["osImage"] = strings.Join(osImage.Items(), ", ")
	return nodeAggregate, nil
}
