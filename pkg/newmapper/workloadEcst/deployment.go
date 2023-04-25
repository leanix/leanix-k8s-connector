package newmapper

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/workload"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"strconv"
	"time"
)

func (m *mapworkload) MapDeploymentsEcst(clusterName string, deployments *appsv1.DeploymentList, services *v1.ServiceList) ([]workload.WorkloadEcst, error) {
	var allDeployments []workload.WorkloadEcst

	for _, deployment := range deployments.Items {
		deploymentService := ""
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		deploymentService = ResolveK8sServiceForK8sDeployment(services, deployment)
		allDeployments = append(allDeployments, m.CreateDeploymentEcst(clusterName, deploymentService, deployment))
	}

	return allDeployments, nil
}

func (m *mapworkload) CreateDeploymentEcst(clusterName string, deploymentService string, deployment appsv1.Deployment) workload.WorkloadEcst {
	var service = ""
	if deploymentService != "" {
		service = deploymentService
	}
	mappedDeployment := workload.WorkloadEcst{
		ClusterName:  clusterName,
		WorkloadType: "deployment",
		WorkloadName: deployment.Name,
		ServiceName:  service,
		Labels:       deployment.ObjectMeta.Labels,
		Timestamp:    deployment.CreationTimestamp.UTC().Format(time.RFC3339),
		Containers: workload.Containers{
			Name:        deployment.Spec.Template.Spec.Containers[0].Name,
			Image:       deployment.Spec.Template.Spec.Containers[0].Image,
			Port:        deployment.Spec.Template.Spec.Containers[0].Ports,
			K8sLimits:   CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Limits),
			K8sRequests: CreateK8sResources(deployment.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
		WorkloadProperties: workload.Properties{
			Replicas:       strconv.FormatInt(int64(deployment.Status.Replicas), 10),
			UpdateStrategy: string(deployment.Spec.Strategy.Type),
		},
	}
	return mappedDeployment
}

func CreateK8sResources(resourceList v1.ResourceList) workload.K8sResources {
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

	return workload.K8sResources{
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
