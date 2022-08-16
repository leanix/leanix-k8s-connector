package iris

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/set"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Mapper interface {
	MapCluster(clusterName string, nodes *v1.NodeList) (ClusterDTO, error)
	MapDeployments(deployments *appsv1.DeploymentList, services *v1.ServiceList, replicaSets *appsv1.ReplicaSetList) ([]models.Deployment, error)
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

// GetCluster MapNodes maps a list of nodes and a given cluster name into a KubernetesObject.
// In the process it aggregates the information from muliple nodes into one cluster object.
func (m *mapper) MapCluster(clusterName string, nodes *v1.NodeList) (ClusterDTO, error) {

	items := nodes.Items
	if len(items) == 0 {
		return ClusterDTO{
			name: clusterName,
		}, nil
	}
	k8sVersion := set.NewStringSet()
	osImage := set.NewStringSet()

	for _, n := range items {
		k8sVersion.Add(n.Status.NodeInfo.KubeletVersion)
		osImage.Add(n.Status.NodeInfo.OSImage)

	}
	return ClusterDTO{
		name:       clusterName,
		k8sVersion: strings.Join(k8sVersion.Items(), ", "),
		nodesCount: len(items),
		osImage:    strings.Join(osImage.Items(), ", "),
	}, nil
}

func (m *mapper) MapDeployments(deployments *appsv1.DeploymentList, services *v1.ServiceList, replicaSets *appsv1.ReplicaSetList) ([]models.Deployment, error) {
	var allDeployments []models.Deployment

	for _, deployment := range deployments.Items {
		deploymentService := ""
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		deploymentReplicaSets := ResolveK8sReplicateSetsForK8sDeployment(replicaSets, deployment)
		deploymentService = ResolveK8sServiceForK8sDeployment(services, deployment)
		allDeployments = append(allDeployments, m.CreateDeployment(deploymentService, deploymentReplicaSets, deployment))
	}

	return allDeployments, nil
}

func (m *mapper) CreateDeployment(deploymentService string, replicaSets []appsv1.ReplicaSet, deployment appsv1.Deployment) models.Deployment {
	lastDeployment := ""
	if len(replicaSets) != 0 {
		lastDeployment = replicaSets[0].CreationTimestamp.UTC().Format(time.RFC3339)
	}

	var service *models.Service = nil
	if deploymentService != "" {
		service = &models.Service{
			Name: deploymentService,
		}
	}

	mappedDeployment := models.Deployment{
		Service:      service,
		Image:        deployment.Spec.Template.Spec.Containers[0].Image,
		Name:         deployment.Name,
		Labels:       deployment.ObjectMeta.Labels,
		Timestamp:    deployment.CreationTimestamp.UTC().Format(time.RFC3339),
		LastDeployed: lastDeployment,
		Properties: models.Properties{
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
		for label, _ := range service.Spec.Selector {
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

func ResolveK8sReplicateSetsForK8sDeployment(replicaSets *appsv1.ReplicaSetList, deployment appsv1.Deployment) []appsv1.ReplicaSet {
	deploymentReplicaSets := make([]appsv1.ReplicaSet, 0)
	for _, replicaSet := range replicaSets.Items {
		sharedLabelsDeployment := map[string]string{}
		sharedLabelsReplicaSet := map[string]string{}
		for label, _ := range replicaSet.Spec.Selector.MatchLabels {
			if _, ok := deployment.Spec.Selector.MatchLabels[label]; ok {
				sharedLabelsDeployment[label] = deployment.Spec.Selector.MatchLabels[label]
				sharedLabelsReplicaSet[label] = replicaSet.Spec.Selector.MatchLabels[label]
			}
		}

		if len(sharedLabelsDeployment) != 0 && len(sharedLabelsReplicaSet) != 0 && reflect.DeepEqual(sharedLabelsDeployment, sharedLabelsReplicaSet) {
			deploymentReplicaSets = append(deploymentReplicaSets, replicaSet)
		}
	}
	sort.SliceStable(deploymentReplicaSets, func(x, y int) bool {
		return deploymentReplicaSets[x].CreationTimestamp.Before(&deploymentReplicaSets[y].CreationTimestamp)
	})
	return deploymentReplicaSets
}
