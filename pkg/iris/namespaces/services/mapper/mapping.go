package mapper

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/namespaces/models"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/set"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type Mapper interface {
	MapCluster(clusterName string, nodes *v1.NodeList) (ClusterDTO, error)
	MapDeploymentsEcst(deployments *appsv1.DeploymentList, services *v1.ServiceList) ([]models.DeploymentEcst, error)
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
	Name       string
	K8sVersion string
	NodesCount int
	OsImage    string
}

// GetCluster MapNodes maps a list of nodes and a given cluster Name into a KubernetesObject.
// In the process it aggregates the information from muliple nodes into one cluster object.
func (m *mapper) MapCluster(clusterName string, nodes *v1.NodeList) (ClusterDTO, error) {

	items := nodes.Items
	if len(items) == 0 {
		return ClusterDTO{
			Name: clusterName,
		}, nil
	}
	k8sVersion := set.NewStringSet()
	osImage := set.NewStringSet()

	for _, n := range items {
		k8sVersion.Add(n.Status.NodeInfo.KubeletVersion)
		osImage.Add(n.Status.NodeInfo.OSImage)

	}
	return ClusterDTO{
		Name:       clusterName,
		K8sVersion: strings.Join(k8sVersion.Items(), ", "),
		NodesCount: len(items),
		OsImage:    strings.Join(osImage.Items(), ", "),
	}, nil
}

func (m *mapper) MapDeploymentsEcst(deployments *appsv1.DeploymentList, services *v1.ServiceList) ([]models.DeploymentEcst, error) {
	var allDeployments []models.DeploymentEcst

	for _, deployment := range deployments.Items {
		deploymentService := ""
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		deploymentService = ResolveK8sServiceForK8sDeployment(services, deployment)
		allDeployments = append(allDeployments, m.CreateDeploymentEcst(deploymentService, deployment))
	}

	return allDeployments, nil
}

func (m *mapper) CreateDeployment(deploymentService string, deployment appsv1.Deployment) models.Deployment {

	mappedDeployment := models.Deployment{
		Service:        &models.Service{Name: deploymentService},
		Image:          deployment.Spec.Template.Spec.Containers[0].Image,
		DeploymentName: deployment.Name,
		Labels:         deployment.ObjectMeta.Labels,
		Timestamp:      deployment.CreationTimestamp.UTC().Format(time.RFC3339),
		Properties: models.DeploymentProperties{
			UpdateStrategy: string(deployment.Spec.Strategy.Type),
			Replicas:       strconv.FormatInt(int64(deployment.Status.Replicas), 10),
			K8sLimits:      CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Limits),
			K8sRequests:    CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
	}
	return mappedDeployment
}

func (m *mapper) CreateDeploymentEcst(deploymentService string, deployment appsv1.Deployment) models.DeploymentEcst {
	var service = ""
	if deploymentService != "" {
		service = deploymentService
	}
	mappedDeployment := models.DeploymentEcst{
		ServiceName:    service,
		Image:          deployment.Spec.Template.Spec.Containers[0].Image,
		DeploymentName: deployment.Name,
		Labels:         deployment.ObjectMeta.Labels,
		Timestamp:      deployment.CreationTimestamp.UTC().Format(time.RFC3339),
		DeploymentProperties: models.DeploymentProperties{
			UpdateStrategy: string(deployment.Spec.Strategy.Type),
			Replicas:       strconv.FormatInt(int64(deployment.Status.Replicas), 10),
			K8sLimits:      CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Limits),
			K8sRequests:    CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
	}
	return mappedDeployment
}

func CreateK8sResources(resourceList v1.ResourceList) models.K8sResources {
	cpu := resourceList[v1.ResourceCPU]
	cpuString := ""
	if !cpu.IsZero() {
		cpuString = cpu.String()
	}

	memory := resourceList[v1.ResourceMemory]
	memoryString := ""
	if !memory.IsZero() {
		memoryString = memory.String()
	}

	return models.K8sResources{
		Cpu:    cpuString,
		Memory: memoryString,
	}
}

func ResolveK8sServiceForK8sDeployment(services *v1.ServiceList, deployment appsv1.Deployment) string {
	deploymentService := ""
	for _, service := range services.Items {
		sharedLabelsDeployment := map[string]string{}
		sharedLabelsService := map[string]string{}
		for label := range service.Spec.Selector {
			if _, ok := deployment.Spec.Selector.MatchLabels[label]; ok {
				sharedLabelsDeployment[label] = deployment.Spec.Selector.MatchLabels[label]
				sharedLabelsService[label] = service.Spec.Selector[label]
			}
		}

		if len(sharedLabelsDeployment) != 0 && len(sharedLabelsService) != 0 && reflect.DeepEqual(sharedLabelsDeployment, sharedLabelsService) {
			deploymentService = service.Name
			break
		}
	}
	return deploymentService
}
