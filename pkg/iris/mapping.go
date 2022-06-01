package iris

import (
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
)

type Mapper interface {
	GetDeployments() ([]DiscoveryItem, error)
	MapDeployments(deployments *appsv1.DeploymentList) []DiscoveryItem
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
	blackListedNamespaces []string,
	runId string) Mapper {

	return &mapper{
		KubernetesApi:         kubernetesApi,
		ClusterName:           clusterName,
		WorkspaceId:           workspaceId,
		BlackListedNamespaces: blackListedNamespaces,
		runId:                 runId,
	}
}

func (m *mapper) GetDeployments() ([]DiscoveryItem, error) {
	namespaces, err := m.KubernetesApi.Namespaces(m.BlackListedNamespaces)
	if err != nil {
		return nil, err
	}
	var allDiscoveryItems []DiscoveryItem
	for _, namespace := range namespaces.Items {
		if blacklistedNS(namespace.Name, m.BlackListedNamespaces) {
			continue
		}
		log.Infof("Fetching Deployments for namespace [%s]", namespace.Name)
		deployments, err := m.KubernetesApi.Deployments(namespace.Name)
		if err != nil {
			return nil, err
		}
		log.Infof("mapping deployments for namespace [%s]", namespace.Name)
		mappedDeployments := m.MapDeployments(deployments)
		allDiscoveryItems = append(allDiscoveryItems, mappedDeployments...)
	}
	log.Info("Fetching deployments for namespaces completed")
	return allDiscoveryItems, nil
}

func (m *mapper) MapDeployments(deployments *appsv1.DeploymentList) []DiscoveryItem {
	var deploymentDiscoveryItems []DiscoveryItem
	for _, deployment := range deployments.Items {
		deploymentItem := NewDeploymentEvent(*m, deployment)
		softwareArtifactItem := NewSoftwareArtifactEvent(*m, deployment)
		deploymentDiscoveryItems = append(deploymentDiscoveryItems, *deploymentItem, *softwareArtifactItem)
	}
	return deploymentDiscoveryItems
}

func blacklistedNS(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
