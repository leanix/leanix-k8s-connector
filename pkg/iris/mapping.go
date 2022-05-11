package iris

import (
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetDeployments(clusterName string, workspaceId string, namespaces *corev1.NamespaceList, kubernetesAPI *kubernetes.API) ([]DiscoveryItem, error) {
	var allDiscoveryItems []DiscoveryItem
	for _, namespace := range namespaces.Items {
		deployments, err := kubernetesAPI.Deployments(namespace.Name)
		if err != nil {
			return nil, err
		}
		mappedDeployments, err := MapDeployments(clusterName, workspaceId, deployments)
		if err != nil {
			return nil, err
		}
		allDiscoveryItems = append(allDiscoveryItems, mappedDeployments...)

	}
	return allDiscoveryItems, nil
}

func MapDeployments(clusterName string, workspaceId string, deployments *appsv1.DeploymentList) ([]DiscoveryItem, error) {
	var deploymentDiscoveryItems []DiscoveryItem
	for _, deployment := range deployments.Items {
		deployment.ClusterName = clusterName
		deploymentItem := DiscoveryItem{
			ID:      deployment.Namespace + ":" + deployment.Name + "-" + deployment.ClusterName,
			Scope:   "workspace/" + workspaceId,
			Type:    "leanix.vsm.item-discovered.runtimeObject",
			Source:  "kubernetes/" + clusterName,
			Time:    deployment.CreationTimestamp.String(),
			Subject: "deployment/" + deployment.Name,
			Data:    deployment,
		}
		var DeploymentData = make(map[string]interface{})

		DeploymentData["clusterName"] = clusterName
		DeploymentData["name"] = deployment.Namespace + ":" + deployment.Name
		DeploymentData["type"] = "namespaceBased"
		softwareArtifactItem := DiscoveryItem{
			ID:      deployment.Namespace + "_" + deployment.Name,
			Scope:   "workspace/" + workspaceId,
			Type:    "leanix.vsm.item-discovered.kubernetesService",
			Source:  "kubernetes/" + clusterName,
			Time:    deployment.CreationTimestamp.String(),
			Subject: "softwareArtifact/" + deployment.Name,
			Data:    DeploymentData,
		}
		deploymentDiscoveryItems = append(deploymentDiscoveryItems, deploymentItem, softwareArtifactItem)
	}
	return deploymentDiscoveryItems, nil
}
