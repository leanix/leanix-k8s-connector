package mapper

import (
	"reflect"
	"strconv"
	"time"

	"github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func (m *MappedWorkload) MapDeploymentsEcst(cluster models.Cluster, deployments *appsv1.DeploymentList, services *v1.ServiceList) ([]models.Data, error) {
	var allDeployments []models.Data

	for _, deployment := range deployments.Items {
		deploymentService := ""
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		deploymentService = ResolveK8sServiceForK8sDeployment(services, deployment)
		allDeployments = append(allDeployments, m.CreateDeploymentEcst(cluster, deploymentService, deployment))
	}

	return allDeployments, nil
}

func (m *MappedWorkload) CreateDeploymentEcst(cluster models.Cluster, deploymentService string, deployment appsv1.Deployment) models.Data {
	var service = ""
	if deploymentService != "" {
		service = deploymentService
	}
	mappedDeployment := models.Data{
		Workload: models.Workload{
			Name:         deployment.Name,
			WorkloadType: "deployment",
			Labels:       deployment.ObjectMeta.Labels,
			WorkloadProperties: models.WorkloadProperties{
				Replicas:       strconv.FormatInt(int64(deployment.Status.Replicas), 10),
				UpdateStrategy: string(deployment.Spec.Strategy.Type),
				Containers: models.Containers{
					Name:        deployment.Spec.Template.Spec.Containers[0].Name,
					Image:       deployment.Spec.Template.Spec.Containers[0].Image,
					Port:        deployment.Spec.Template.Spec.Containers[0].Ports,
					K8sLimits:   CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Limits),
					K8sRequests: CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Requests),
				},
			},
		},
		Cluster: models.Cluster{
			Name: cluster.Name,
			Os:   cluster.Os,
		},
		ServiceName:   service,
		NamespaceName: deployment.Namespace,
		Timestamp:     deployment.CreationTimestamp.UTC().Format(time.RFC3339),
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
	if deployment.Spec.Selector != nil {

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
	} else {
		logger.Infof("Deployment %s has no selector labels", deployment.Name)
	}
	return deploymentService
}
