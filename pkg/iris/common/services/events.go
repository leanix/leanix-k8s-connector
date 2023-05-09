package services

import (
	"encoding/json"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
	namespace "github.com/leanix/leanix-k8s-connector/pkg/iris/namespaces/models"
	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
)

func FilterForDeletedItems(oldResultMap map[string]models.DiscoveryEvent, workspaceId string, configId string, runId string) ([]models.DiscoveryEvent, error) {
	deletedEvents := make([]models.DiscoveryEvent, 0)
	for _, oldItem := range oldResultMap {
		if oldItem.HeaderProperties.Class == models.EventClassNamespace {
			mappedData, err := ParseNamespaceData(oldItem)
			if err != nil {
				return nil, err
			}
			deletedEvent := namespace.CreateEcstDiscoveryEvent(models.EventTypeChange, models.EventActionDeleted, *mappedData, workspaceId, configId)
			deletedEvents = append(deletedEvents, deletedEvent)
		} else if oldItem.HeaderProperties.Class == models.EventClassWorkload {
			mappedData, err := ParseWorkloadData(oldItem)
			if err != nil {
				return nil, err
			}
			deletedEvent := workload.CreateEcstDiscoveryEvent(models.EventTypeChange, models.EventActionDeleted, *mappedData, runId, workspaceId, configId)
			deletedEvents = append(deletedEvents, deletedEvent)
		}
	}
	return deletedEvents, nil
}

func ParseNamespaceData(oldItem models.DiscoveryEvent) (*namespace.Data, error) {
	dataString, err := json.Marshal(oldItem.Body.State.Data)
	if err != nil {
		return nil, err
	}
	var mappedData namespace.Data
	err = json.Unmarshal(dataString, &mappedData)
	if err != nil {
		return nil, err
	}
	return &mappedData, nil
}

func ParseWorkloadData(oldItem models.DiscoveryEvent) (*workload.Data, error) {
	dataString, err := json.Marshal(oldItem.Body.State.Data)
	if err != nil {
		return nil, err
	}
	var mappedData workload.Data
	err = json.Unmarshal(dataString, &mappedData)
	if err != nil {
		return nil, err
	}
	return &mappedData, nil
}
