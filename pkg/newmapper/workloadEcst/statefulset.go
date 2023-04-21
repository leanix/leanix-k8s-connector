package newmapper

import (
	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/models/workload"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
)

func (m *mapworkload) MapStatefulSetsEcst(clusterName string, statefulSets *appsv1.StatefulSetList, services *v1.ServiceList) ([]workload.WorkloadEcst, error) {
	var allStatefulSets []workload.WorkloadEcst

	for _, statefulSet := range statefulSets.Items {
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		statefulSetService := ResolveK8sServiceForK8sStatefulSet(services, statefulSet)
		mappedStatefulSet := m.CreateStatefulSetEcst(clusterName, statefulSet, statefulSetService)
		allStatefulSets = append(allStatefulSets, mappedStatefulSet)
	}

	return allStatefulSets, nil
}

// CreateStatefulSetEcst create a data object that contains name, labels, StatefulSet properties and more
func (m *mapworkload) CreateStatefulSetEcst(clusterName string, statefulSet appsv1.StatefulSet, service string) workload.WorkloadEcst {
	mappedDeployment := workload.WorkloadEcst{
		ClusterName:  clusterName,
		WorkloadType: "cronjob",
		WorkloadName: statefulSet.Name,
		ServiceName:  service,
		Containers: workload.Containers{
			Name:        statefulSet.Spec.Template.Spec.Containers[0].Name,
			Image:       strings.Split(statefulSet.Spec.Template.Spec.Containers[0].Image, ":")[0],
			Port:        statefulSet.Spec.Template.Spec.Containers[0].Ports,
			K8sLimits:   CreateK8sResources(statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits),
			K8sRequests: CreateK8sResources(statefulSet.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
		WorkloadProperties: workload.Properties{
			Replicas:       string(statefulSet.Status.Replicas),
			Labels:         statefulSet.ObjectMeta.Labels,
			UpdateStrategy: string(statefulSet.Spec.UpdateStrategy.Type),
			Timestamp:      statefulSet.CreationTimestamp.UTC().Format(time.RFC3339),
		},
	}
	return mappedDeployment
}

func ResolveK8sServiceForK8sStatefulSet(services *v1.ServiceList, statefulSet appsv1.StatefulSet) string {
	statefulSetService := ""
	for _, service := range services.Items {
		sharedLabelsStatefulSet := map[string]string{}
		sharedLabelsService := map[string]string{}
		for label := range service.Spec.Selector {
			if _, ok := statefulSet.Spec.Selector.MatchLabels[label]; ok {
				sharedLabelsStatefulSet[label] = statefulSet.Spec.Selector.MatchLabels[label]
				sharedLabelsService[label] = service.Spec.Selector[label]
			}
		}

		if len(sharedLabelsStatefulSet) != 0 && len(sharedLabelsService) != 0 && reflect.DeepEqual(sharedLabelsStatefulSet, sharedLabelsService) {
			statefulSetService = service.Name
			break
		}
	}
	return statefulSetService
}
