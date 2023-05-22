package mapper

import (
	"reflect"
	"strings"
	"time"

	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func (m *mapworkload) MapDaemonSetsEcst(clusterName string, daemonSets *appsv1.DaemonSetList, services *v1.ServiceList) ([]workload.Workload, error) {
	var allDaemonSets []workload.Workload

	for _, daemonSet := range daemonSets.Items {
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		daemonSetService := ResolveK8sServiceForK8sDaemonSet(services, daemonSet)
		mappedDaemonSet := m.CreateDaemonSetEcst(clusterName, daemonSet, daemonSetService)
		allDaemonSets = append(allDaemonSets, mappedDaemonSet)
	}

	return allDaemonSets, nil
}

// CreateDaemonSetEcst create a data object that contains name, labels, DaemonSet properties and more
func (m *mapworkload) CreateDaemonSetEcst(clusterName string, daemonSet appsv1.DaemonSet, service string) workload.Workload {
	mappedDeployment := workload.Workload{
		ClusterName:  clusterName,
		WorkloadType: "daemonSet",
		WorkloadName: daemonSet.Name,
		ServiceName:  service,
		Labels:       daemonSet.ObjectMeta.Labels,
		Timestamp:    daemonSet.CreationTimestamp.UTC().Format(time.RFC3339),
		Containers: workload.Containers{
			Name:        daemonSet.Spec.Template.Spec.Containers[0].Name,
			Image:       strings.Split(daemonSet.Spec.Template.Spec.Containers[0].Image, ":")[0],
			Port:        daemonSet.Spec.Template.Spec.Containers[0].Ports,
			K8sLimits:   CreateK8sResources(daemonSet.Spec.Template.Spec.Containers[0].Resources.Limits),
			K8sRequests: CreateK8sResources(daemonSet.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
		WorkloadProperties: workload.Properties{
			UpdateStrategy: string(daemonSet.Spec.UpdateStrategy.Type),
		},
	}
	return mappedDeployment
}

func ResolveK8sServiceForK8sDaemonSet(services *v1.ServiceList, daemonSet appsv1.DaemonSet) string {
	daemonSetService := ""
	if daemonSet.Spec.Selector != nil {
		for _, service := range services.Items {
			sharedLabelsStatefulSet := map[string]string{}
			sharedLabelsService := map[string]string{}
			for label := range service.Spec.Selector {
				if _, ok := daemonSet.Spec.Selector.MatchLabels[label]; ok {
					sharedLabelsStatefulSet[label] = daemonSet.Spec.Selector.MatchLabels[label]
					sharedLabelsService[label] = service.Spec.Selector[label]
				}
			}

			if len(sharedLabelsStatefulSet) != 0 && len(sharedLabelsService) != 0 && reflect.DeepEqual(sharedLabelsStatefulSet, sharedLabelsService) {
				daemonSetService = service.Name
				break
			}
		}
	} else {
		logger.Infof("DaemonSet %s has no selector labels", daemonSet.Name)
	}
	return daemonSetService
}