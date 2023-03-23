package iris

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/pkg/errors"
)

type EventProducer interface {
	PostLegacyResults(data []models.DiscoveryData) error
	ProcessResults(data []models.Data, oldData []models.DiscoveryEvent, configId string) error
	PostStatus(status []byte) error
	FilterForChangedItems(newData map[string]models.Data, oldData map[string]models.DiscoveryEvent, configId string) ([]models.DiscoveryEvent, []models.DiscoveryEvent, map[string]models.DiscoveryEvent, error)
}

type eventProducer struct {
	irisApi     API
	runId       string
	workspaceId string
}

func NewEventProducer(irisApi API, runId string, workspaceId string) EventProducer {
	return &eventProducer{
		irisApi:     irisApi,
		runId:       runId,
		workspaceId: workspaceId,
	}
}

func (p *eventProducer) PostLegacyResults(data []models.DiscoveryData) error {
	events := make([]models.DiscoveryItem, 0)
	for _, item := range data {
		// Metadata for ECST Discovery Item
		scope := fmt.Sprintf("workspace/%s", p.workspaceId)
		source := fmt.Sprintf("kubernetes/%s#%s", item.Cluster.Name, p.runId)
		events = append(events, CreateDiscoveryItem(item.Cluster, source, scope))

	}
	if len(events) == 0 {
		return nil
	}
	scannedObjectsByte, err := storage.Marshal(events)
	if err != nil {
		return errors.Wrap(err, "Marshall scanned legacy services")
	}
	return p.irisApi.PostResults(scannedObjectsByte)
}

func (p *eventProducer) ProcessResults(data []models.Data, oldData []models.DiscoveryEvent, configId string) error {
	created, updated, deleted, err := p.createECSTEvents(data, oldData, configId)
	if err != nil {
		return err
	}
	ecstEvents := append(created, updated...)
	ecstEvents = append(ecstEvents, deleted...)
	if len(ecstEvents) == 0 {
		return nil
	}
	scannedEcstObjectsByte, err := storage.Marshal(ecstEvents)
	if err != nil {
		return errors.Wrap(err, "Marshall scanned ECST services")
	}

	return p.irisApi.PostEcstResults(scannedEcstObjectsByte)
}

func (p *eventProducer) PostStatus(status []byte) error {
	return p.irisApi.PostStatus(status)
}

func (p *eventProducer) createECSTEvents(data []models.Data, oldData []models.DiscoveryEvent, configId string) ([]models.DiscoveryEvent, []models.DiscoveryEvent, []models.DiscoveryEvent, error) {
	deletedEvents := make([]models.DiscoveryEvent, 0)
	resultMap := p.createItemMap(data, configId)
	oldResultMap := p.createOldItemMap(oldData)
	for id, r := range oldResultMap {
		logger.Infof("Current db item id:%s, scope: %s, ClusterName: %s, namespace:%s", id, r.HeaderProperties.Scope, r.Body.State.Data.Cluster.Name, r.Body.State.Data.Cluster.Namespace)
	}

	createdEvents, updatedEvents, oldResultMap, err := p.FilterForChangedItems(resultMap, oldResultMap, configId)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create DELETED events
	for _, oldItem := range oldResultMap {
		deletedEvent := CreateEcstDiscoveryEvent(EventTypeChange, EventActionDeleted, oldItem.Body.State.Data, p.runId, p.workspaceId, configId)
		logger.Infof("DELETE event id:%s, scope: %s, ClusterName: %s, namespace:%s", deletedEvent.HeaderProperties.Id, deletedEvent.HeaderProperties.Scope, deletedEvent.Body.State.Data.Cluster.Name, deletedEvent.Body.State.Data.Cluster.Namespace)
		deletedEvents = append(deletedEvents, deletedEvent)
	}
	return createdEvents, updatedEvents, deletedEvents, nil

}

func (p *eventProducer) createItemMap(data []models.Data, configId string) map[string]models.Data {
	resultMap := map[string]models.Data{}
	for _, item := range data {
		// Build unique string hash for discoveryItem
		id := GenerateId(p.workspaceId, configId, item)
		resultMap[id] = item
	}
	return resultMap
}

func (p *eventProducer) createOldItemMap(data []models.DiscoveryEvent) map[string]models.DiscoveryEvent {
	resultMap := map[string]models.DiscoveryEvent{}
	for _, item := range data {
		// Use id hash from iris to as a key
		resultMap[item.HeaderProperties.Id] = item
	}
	return resultMap
}

func GenerateHash(toBeHashed interface{}) (string, error) {
	serialized, err := json.Marshal(toBeHashed)
	if err != nil {
		return "", err
	}
	hashed := sha256.Sum256(serialized)
	return hex.EncodeToString(hashed[:]), nil

}

func (p *eventProducer) FilterForChangedItems(newData map[string]models.Data, oldData map[string]models.DiscoveryEvent, configId string) ([]models.DiscoveryEvent, []models.DiscoveryEvent, map[string]models.DiscoveryEvent, error) {
	updated := make([]models.DiscoveryEvent, 0)
	created := make([]models.DiscoveryEvent, 0)
	for id, newItem := range newData {
		// if the current element from the freshly discovered items is not in the old results, create an CREATED event
		if oldItem, ok := oldData[id]; !ok {
			createdEcstDiscoveryEvent := CreateEcstDiscoveryEvent(EventTypeChange, EventActionCreated, newItem, p.runId, p.workspaceId, configId)
			created = append(created, createdEcstDiscoveryEvent)
			logger.Infof("CREATE event id:%s, scope: %s, ClusterName: %s, namespace:%s", createdEcstDiscoveryEvent.HeaderProperties.Id, createdEcstDiscoveryEvent.HeaderProperties.Scope, createdEcstDiscoveryEvent.Body.State.Data.Cluster.Name, createdEcstDiscoveryEvent.Body.State.Data.Cluster.Namespace)
			// if item has been discovered before, check if there are any changes in the new payload
		} else {
			oldItemHash, err := GenerateHash(oldItem.Body.State.Data)
			if err != nil {
				return nil, nil, nil, err
			}
			newItemHash, err := GenerateHash(newItem)
			if err != nil {
				return nil, nil, nil, err
			}

			if oldItemHash != newItemHash {
				log.Printf("old %+v, new %+v", oldItem.Body.State.Data.Cluster, newItem.Cluster)
				updated = append(updated, CreateEcstDiscoveryEvent(EventTypeChange, EventActionUpdated, newItem, p.runId, p.workspaceId, configId))
			}
			// Remove key from oldData results, so we only have the entries inside which shall be deleted
			delete(oldData, id)
		}
	}

	return created, updated, oldData, nil
}
