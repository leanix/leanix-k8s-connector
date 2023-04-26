package services

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
	namespace "github.com/leanix/leanix-k8s-connector/pkg/iris/namespaces/models"
	namespaceEvents "github.com/leanix/leanix-k8s-connector/pkg/iris/namespaces/services/events"
	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	workloadEvents "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/services/events"
)

func FilterForDeletedItems(oldResultMap map[string]models.DiscoveryEvent, workspaceId string, configId string, runId string) ([]models.DiscoveryEvent, error) {
	deletedEvents := make([]models.DiscoveryEvent, 0)
	for _, oldItem := range oldResultMap {
		if oldItem.HeaderProperties.Class == models.EventClassNamespace {
			mappedData, err := namespaceEvents.ParseNamespaceData(oldItem)
			if err != nil {
				return nil, err
			}
			deletedEvent := namespace.CreateEcstDiscoveryEvent(models.EventTypeChange, models.EventActionDeleted, *mappedData, workspaceId, configId)
			deletedEvents = append(deletedEvents, deletedEvent)
		} else if oldItem.HeaderProperties.Class == models.EventClassWorkload {
			mappedData, err := workloadEvents.ParseWorkloadData(oldItem)
			if err != nil {
				return nil, err
			}
			deletedEvent := workload.CreateEcstDiscoveryEvent(models.EventTypeChange, models.EventActionDeleted, *mappedData, runId, workspaceId, configId)
			deletedEvents = append(deletedEvents, deletedEvent)
		}
	}
	return deletedEvents, nil
}
