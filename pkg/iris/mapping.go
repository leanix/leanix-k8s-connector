package iris

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/set"
	"strconv"
	"strings"
	"time"
)

type Mapper interface {
	GetCluster(clusterName string, kubernetesAPI *kubernetes.API) (*ClusterDTO, error)
	GetDeployments(namespace string, kubernetesAPI *kubernetes.API) ([]models.Deployment, error)
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

type ClusterDTO struct {
	name       string
	k8sVersion string
	nodesCount int
	osImage    string
}

// MapNodes maps a list of nodes and a given cluster name into a KubernetesObject.
// In the process it aggregates the information from muliple nodes into one cluster object.
func (m *mapper) GetCluster(clusterName string, kubernetesAPI *kubernetes.API) (*ClusterDTO, error) {

	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		return nil, err
	}
	items := nodes.Items
	if len(items) == 0 {
		return &ClusterDTO{
			name: clusterName,
		}, nil
	}
	k8sVersion := set.NewStringSet()
	osImage := set.NewStringSet()

	for _, n := range items {
		k8sVersion.Add(n.Status.NodeInfo.KubeletVersion)
		osImage.Add(n.Status.NodeInfo.OSImage)

	}
	return &ClusterDTO{
		name:       clusterName,
		k8sVersion: strings.Join(k8sVersion.Items(), ", "),
		nodesCount: len(items),
		osImage:    strings.Join(osImage.Items(), ", "),
	}, nil
}

func (m *mapper) GetDeployments(namespace string, kubernetesAPI *kubernetes.API) ([]models.Deployment, error) {
	var allDeployments []models.Deployment
	deployments, err := kubernetesAPI.Deployments(namespace)
	if err != nil {
		return nil, err
	}

	services, err := kubernetesAPI.Services(namespace)
	if err != nil {
		return nil, err
	}
	allServices := set.NewStringSet()
	for _, service := range services.Items {
		allServices.Add(service.Name)
	}
	for _, deployment := range deployments.Items {
		deploymentService := ""
		if allServices.Contains(deployment.Name) {
			deploymentService = deployment.Name
		}
		limitCpu := deployment.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"]
		limitMemory := deployment.Spec.Template.Spec.Containers[0].Resources.Limits["memory"]
		requestCpu := deployment.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"]
		requestMemory := deployment.Spec.Template.Spec.Containers[0].Resources.Requests["memory"]
		mappedDeployment := models.Deployment{
			Service: models.Service{
				Name: deploymentService,
			},
			Name:      deployment.Name,
			Labels:    deployment.ObjectMeta.Labels,
			Timestamp: deployment.CreationTimestamp.UTC().Format(time.RFC3339),
			Properties: models.Properties{
				UpdateStrategy: string(deployment.Spec.Strategy.Type),
				Replicas:       strconv.FormatInt(int64(deployment.Status.Replicas), 10),
				K8sLimits: models.K8sLimits{
					Cpu:    limitCpu.String(),
					Memory: limitMemory.String(),
				},
				K8sRequests: models.K8sRequests{
					Cpu:    requestCpu.String(),
					Memory: requestMemory.String(),
				},
			},
		}
		allDeployments = append(allDeployments, mappedDeployment)
	}

	return allDeployments, nil
}
