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

func (m *workloadMapper) MapStatefulSetsEcst(cluster workload.Cluster, statefulSets *appsv1.StatefulSetList, services *v1.ServiceList) ([]workload.Data, error) {
	var allStatefulSets []workload.Data

	for _, statefulSet := range statefulSets.Items {
		// Check if any service has the exact same selector labels and use this as the service related to the deployment
		statefulSetService := ResolveK8sServiceForK8sStatefulSet(services, statefulSet)
		mappedStatefulSet := m.CreateStatefulSetEcst(cluster, statefulSet, statefulSetService)
		allStatefulSets = append(allStatefulSets, mappedStatefulSet)
	}

	return allStatefulSets, nil
}

// CreateStatefulSetEcst create a data object that contains name, labels, StatefulSet properties and more
func (m *workloadMapper) CreateStatefulSetEcst(cluster workload.Cluster, statefulSet appsv1.StatefulSet, service string) workload.Data {
	mappedDeployment := workload.Data{
		Workload: workload.Workload{
			Name:         statefulSet.Name,
			WorkloadType: "statefulSet",
			Labels:       statefulSet.ObjectMeta.Labels,
			WorkloadProperties: workload.WorkloadProperties{
				Replicas:       string(statefulSet.Status.Replicas),
				UpdateStrategy: string(statefulSet.Spec.UpdateStrategy.Type),
				Containers: workload.Containers{
					Name:        statefulSet.Spec.Template.Spec.Containers[0].Name,
					Image:       strings.Split(statefulSet.Spec.Template.Spec.Containers[0].Image, ":")[0],
					Port:        statefulSet.Spec.Template.Spec.Containers[0].Ports,
					K8sLimits:   CreateK8sResources(statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits),
					K8sRequests: CreateK8sResources(statefulSet.Spec.Template.Spec.Containers[0].Resources.Requests),
				},
			},
		},
		Cluster: workload.Cluster{
			Name:       cluster.Name,
			OsImage:    cluster.OsImage,
			NoOfNodes:  cluster.NoOfNodes,
			K8sVersion: cluster.K8sVersion,
		},
		ServiceName:   service,
		NamespaceName: statefulSet.Namespace,
		Timestamp:     statefulSet.CreationTimestamp.UTC().Format(time.RFC3339),
	}
	return mappedDeployment
}

func ResolveK8sServiceForK8sStatefulSet(services *v1.ServiceList, statefulSet appsv1.StatefulSet) string {
	statefulSetService := ""
	if statefulSet.Spec.Selector != nil {

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
	} else {
		logger.Infof("StatefulSet %s has no selector labels", statefulSet.Name)
	}
	return statefulSetService
}
