package iris

import (
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/pkg/errors"
)

type EventProducer interface {
	PostLegacyResults(data []models.DiscoveryData) error
	ProcessResults(data []models.Data, oldData []models.IrisResultItem) error
	PostStatus(status []byte) error
}

type eventProducer struct {
	irisApi     API
	runId       string
	configId    string
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
	scannedObjectsByte, err := storage.Marshal(events)
	if err != nil {
		return errors.Wrap(err, "Marshall scanned legacy services")
	}
	return p.irisApi.PostResults(scannedObjectsByte)
}

func (p *eventProducer) ProcessResults(data []models.Data, oldData []models.IrisResultItem) error {
	created, updated, deleted, err := p.createECSTEvents(data, oldData)
	if err != nil {
		return err
	}
	ecstEvents := append(created, updated...)
	ecstEvents = append(ecstEvents, deleted...)
	scannedEcstObjectsByte, err := storage.Marshal(ecstEvents)
	if err != nil {
		return errors.Wrap(err, "Marshall scanned ECST services")
	}
	return p.irisApi.PostEcstResults(scannedEcstObjectsByte)
}

func (p *eventProducer) PostStatus(status []byte) error {
	return p.irisApi.PostStatus(status)
}

func (p *eventProducer) createECSTEvents(data []models.Data, oldData []models.IrisResultItem) ([]models.DiscoveryEvent, []models.DiscoveryEvent, []models.DiscoveryEvent, error) {
	created := make([]models.DiscoveryEvent, 0)
	updated := make([]models.DiscoveryEvent, 0)
	deleted := make([]models.DiscoveryEvent, 0)
	resultMap := p.createItemMap(data)
	oldResultMap := p.createOldItemMap(oldData)

	data, createdEvents, err := p.filterForNewDiscoveries(data, oldData)
	if err != nil {
		return nil, nil, nil, err
	}

	return created, updated, deleted, nil

}

func (p *eventProducer) createItemMap(data []models.Data) map[string]models.Data {
	resultMap := map[string]models.Data{}
	for _, item := range data {
		// Build unique string for discoveryItem
		scope := fmt.Sprintf("workspace/%s/configuration/%s", p.workspaceId, p.configId)
		source := fmt.Sprintf("kubernetes/%s/discoveryItem/service/kubernetes", item.Cluster.Name)
		resultMap[fmt.Sprintf("%s/%s", scope, source)] = item
	}
	return resultMap
}

func (p *eventProducer) createOldItemMap(data []models.IrisResultItem) map[string]models.IrisResultItem {
	resultMap := map[string]models.IrisResultItem{}
	for _, item := range data {
		// Build unique string for discoveryItem
		scope := fmt.Sprintf("workspace/%s/configuration/%s", p.workspaceId, p.configId)
		source := fmt.Sprintf("kubernetes/%s/discoveryItem/service/kubernetes", item.Data[""])
		resultMap[fmt.Sprintf("%s/%s", scope, source)] = item
	}
	return resultMap
}

func (p *eventProducer) filterForNewDiscoveries(data []models.Data, oldData []models.IrisResultItem) ([]models.Data, []models.DiscoveryEvent, error) {
	created := make([]models.DiscoveryEvent, 0)
	filteredData := make([]models.Data, 0)
	for _, item := range data {
		scope := fmt.Sprintf("workspace/%s/configuration/%s", p.workspaceId, p.configId)
		source := fmt.Sprintf("kubernetes/%s#%s/", item.Cluster.Name, p.runId)
		created = append(created, CreateEcstDiscoveryEvent(EVENT_ACTION_CREATED, "", item, source, scope))

		// TODO remove
	}

	return filteredData, created, nil
}
