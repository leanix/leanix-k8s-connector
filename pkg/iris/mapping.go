package iris

import (
	"fmt"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
)

type Mapper interface {
	GetDeployments() ([]DiscoveryItem, error)
	MapDeployments(deployments *appsv1.DeploymentList) ([]DiscoveryItem, error)
}

type mapper struct {
	KubernetesApi         *kubernetes.API
	ClusterName           string
	WorkspaceId           string
	BlackListedNamespaces []string
	runId                 string
}

func NewMapper(
	kubernetesApi *kubernetes.API,
	clusterName string,
	workspaceId string,
	blackListedNamespaces []string, runId string) (Mapper, error) {

	return &mapper{
		KubernetesApi:         kubernetesApi,
		ClusterName:           clusterName,
		WorkspaceId:           workspaceId,
		BlackListedNamespaces: blackListedNamespaces,
		runId:                 runId,
	}, nil
}

func (m *mapper) GetDeployments() ([]DiscoveryItem, error) {
	namespaces, err := m.KubernetesApi.Namespaces(m.BlackListedNamespaces)
	if err != nil {
		return nil, err
	}
	var allDiscoveryItems []DiscoveryItem
	for _, namespace := range namespaces.Items {
		deployments, err := m.KubernetesApi.Deployments(namespace.Name)
		if err != nil {
			return nil, err
		}
		mappedDeployments, err := m.MapDeployments(deployments)
		if err != nil {
			return nil, err
		}
		allDiscoveryItems = append(allDiscoveryItems, mappedDeployments...)

	}
	return allDiscoveryItems, nil
}

func (m *mapper) MapDeployments(deployments *appsv1.DeploymentList) ([]DiscoveryItem, error) {
	var deploymentDiscoveryItems []DiscoveryItem
	for _, deployment := range deployments.Items {
		deployment.ClusterName = m.ClusterName
		id := fmt.Sprintf("%s:%s-%s", deployment.Namespace, deployment.Name, deployment.ClusterName)
		Type := "leanix.vsm.item-discovered.runtimeObject"
		Subject := fmt.Sprintf("deployment/%s", deployment.Name)

		// common attributes
		scope := fmt.Sprintf("workspace/%s", m.WorkspaceId)
		Source := fmt.Sprintf("kubernetes/%s#%s", m.ClusterName, m.runId)
		Time := deployment.CreationTimestamp.String()

		deploymentItem := GenerateDiscoveryItem(id, scope, Type, Source, Time, Subject, deployment)

		var DeploymentData = make(map[string]interface{})

		DeploymentData["clusterName"] = m.ClusterName
		DeploymentData["name"] = deployment.Namespace + ":" + deployment.Name
		DeploymentData["type"] = "namespaceBased"
		id = fmt.Sprintf("%s-%s", deployment.Namespace, deployment.Name)
		Type = "leanix.vsm.item-discovered.kubernetesService"
		Subject = fmt.Sprintf("softwareArtifact/%s", deployment.Name)
		softwareArtifactItem := GenerateDiscoveryItem(id, scope, Type, Source, Time, Subject, DeploymentData)

		deploymentDiscoveryItems = append(deploymentDiscoveryItems, *deploymentItem, *softwareArtifactItem)
	}
	return deploymentDiscoveryItems, nil
}
