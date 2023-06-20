package events

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
	common "github.com/leanix/leanix-k8s-connector/pkg/iris/common/services"
	workload "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	"github.com/pkg/errors"
)

type WorkloadEventProducer interface {
	ProcessWorkloads(data []workload.Data, oldData []models.DiscoveryEvent, configId string) error
	PostStatus(status []byte) error
	FilterForChangedItems(newData map[string]workload.Data, oldData map[string]models.DiscoveryEvent, configId string) ([]models.DiscoveryEvent, []models.DiscoveryEvent, map[string]models.DiscoveryEvent, error)
}

type workloadEventProducer struct {
	irisApi     common.IrisApi
	runId       string
	workspaceId string
}

func NewEventWorkloadProducer(irisApi common.IrisApi, runId string, workspaceId string) WorkloadEventProducer {
	return &workloadEventProducer{
		irisApi:     irisApi,
		runId:       runId,
		workspaceId: workspaceId,
	}
}

func (p *workloadEventProducer) ProcessWorkloads(data []workload.Data, oldData []models.DiscoveryEvent, configId string) error {
	created, updated, deleted, err := p.CreateECSTWorkloadEvents(data, oldData, configId)
	if err != nil {
		return err
	}
	ecstEvents := append(created, updated...)
	ecstEvents = append(ecstEvents, deleted...)
	if len(ecstEvents) == 0 {
		return nil
	}
	scannedEcstObjectsByte, err := json.Marshal(ecstEvents)
	if err != nil {
		return errors.Wrap(err, "Marshall scanned ECST services")
	}

	return p.irisApi.PostEcstResults(scannedEcstObjectsByte)
}

func (p *workloadEventProducer) PostStatus(status []byte) error {
	return p.irisApi.PostStatus(status)
}

func (p *workloadEventProducer) CreateECSTWorkloadEvents(data []workload.Data, oldData []models.DiscoveryEvent, configId string) ([]models.DiscoveryEvent, []models.DiscoveryEvent, []models.DiscoveryEvent, error) {
	resultMap := p.createItemMap(data, configId)
	oldResultMap := p.createOldItemMap(oldData)

	createdEvents, updatedEvents, oldResultMap, err := p.FilterForChangedItems(resultMap, oldResultMap, configId)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create DELETED events
	deletedEvents, err := common.FilterForDeletedItems(oldResultMap, p.workspaceId, configId, p.runId)
	if err != nil {
		return nil, nil, nil, err
	}
	return createdEvents, updatedEvents, deletedEvents, nil

}

func (p *workloadEventProducer) createItemMap(workloads []workload.Data, configId string) map[string]workload.Data {
	resultMap := map[string]workload.Data{}
	for _, item := range workloads {
		// Build unique string hash for discoveryItem
		id := workload.GenerateId(p.workspaceId, configId, item)
		resultMap[id] = item
	}
	return resultMap
}

func (p *workloadEventProducer) createOldItemMap(data []models.DiscoveryEvent) map[string]models.DiscoveryEvent {
	resultMap := map[string]models.DiscoveryEvent{}
	for _, item := range data {
		// Use id hash from iris to as a key
		resultMap[item.HeaderProperties.Id] = item
	}
	return resultMap
}

func GenerateHashWorkload(toBeHashed interface{}) (string, error) {
	serialized, err := json.Marshal(toBeHashed)
	if err != nil {
		return "", err
	}
	hashed := sha256.Sum256(serialized)
	return hex.EncodeToString(hashed[:]), nil

}

func (p *workloadEventProducer) FilterForChangedItems(newData map[string]workload.Data, oldData map[string]models.DiscoveryEvent, configId string) ([]models.DiscoveryEvent, []models.DiscoveryEvent, map[string]models.DiscoveryEvent, error) {
	updated := make([]models.DiscoveryEvent, 0)
	created := make([]models.DiscoveryEvent, 0)
	for id, newItem := range newData {
		// if the current element from the freshly discovered items is not in the old results, create an CREATED event
		if oldItem, ok := oldData[id]; !ok {
			createdEcstDiscoveryEvent := workload.CreateEcstDiscoveryEvent(models.EventTypeChange, models.EventActionCreated, id, newItem, p.runId, p.workspaceId, configId)
			created = append(created, createdEcstDiscoveryEvent)
			// if item has been discovered before, check if there are any changes in the new payload
		} else {
			oldItemHash, err := GenerateHashWorkload(oldItem.Body.State.Data)
			if err != nil {
				return nil, nil, nil, err
			}
			newItemHash, err := GenerateHashWorkload(newItem)
			if err != nil {
				return nil, nil, nil, err
			}

			if oldItemHash != newItemHash {
				updated = append(updated, workload.CreateEcstDiscoveryEvent(models.EventTypeChange, models.EventActionUpdated, id, newItem, p.runId, p.workspaceId, configId))
			}
			// Remove key from oldData results, so we only have the entries inside which shall be deleted
			delete(oldData, id)
		}
	}

	return created, updated, oldData, nil
}
