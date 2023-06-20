package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	time2 "time"

	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
)

// ECST Discovery Items
type EcstEventBuilder interface {
	Header(header models.HeaderProperties) EcstEventBuilder
	Body(body models.DiscoveryBody) EcstEventBuilder

	Build() models.DiscoveryEvent
}

type CommandBuilder interface {
	Header(header models.CommandProperties) CommandBuilder
	Body(body models.CommandBody) CommandBuilder

	Build() models.CommandEvent
}

type ecstEventBuilder struct {
	header models.HeaderProperties
	body   models.DiscoveryBody
}

type EcstWorkloadEventBuilder interface {
	Header(header models.HeaderProperties) EcstEventBuilder
	Body(body models.DiscoveryBody) EcstEventBuilder

	Build() models.DiscoveryEvent
}

type commandBuilder struct {
	header models.CommandProperties
	body   models.CommandBody
}

// NewEcstBuilder Discovery Event Builder
func NewEcstBuilder() EcstEventBuilder {
	return &ecstEventBuilder{}
}

func (eb *ecstEventBuilder) Header(header models.HeaderProperties) EcstEventBuilder {
	eb.header = header
	return eb
}

func (eb *ecstEventBuilder) Body(body models.DiscoveryBody) EcstEventBuilder {
	eb.body = body
	return eb
}

func (eb *ecstEventBuilder) Build() models.DiscoveryEvent {
	body := &models.DiscoveryBody{
		State: eb.body.State,
	}

	header := &models.HeaderProperties{
		Class:  eb.header.Class,
		Type:   eb.header.Type,
		Scope:  eb.header.Scope,
		Id:     eb.header.Id,
		Action: eb.header.Action,
	}
	return models.DiscoveryEvent{
		HeaderProperties: *header,
		Body:             *body,
	}
}

// Command Event Builder
func NewCommand() CommandBuilder {
	return &commandBuilder{}
}

func (cb *commandBuilder) Header(header models.CommandProperties) CommandBuilder {
	cb.header = header
	return cb
}

func (cb *commandBuilder) Body(body models.CommandBody) CommandBuilder {
	cb.body = body
	return cb
}

func (cb *commandBuilder) Build() models.CommandEvent {
	header := &models.CommandProperties{
		Type:   cb.header.Type,
		Scope:  cb.header.Scope,
		Action: cb.header.Action,
	}
	body := &models.CommandBody{}

	return models.CommandEvent{
		Properties: *header,
		Body:       *body,
	}
}

// CreateEcstDiscoveryEvent ECST Discovery Items
func CreateEcstDiscoveryEvent(eventType string, changeAction string, id string, data Data, runId string, workspaceId string, configId string) models.DiscoveryEvent {
	// Metadata for the event
	time := time2.Now().Format(time2.RFC3339)

	var header models.HeaderProperties
	if eventType == models.EventTypeChange {
		header = models.HeaderProperties{
			Id:     id,
			Scope:  fmt.Sprintf(models.EventScopeFormat, workspaceId, configId),
			Class:  models.EventClassWorkload,
			Type:   eventType,
			Action: changeAction,
		}
	} else {
		header = models.HeaderProperties{
			Id:    id,
			Scope: fmt.Sprintf(models.EventScopeFormat, workspaceId, configId),
			Class: models.EventClassWorkload,
			Type:  eventType,
		}
	}

	body := models.DiscoveryBody{
		State: models.State{
			Name:           fmt.Sprintf("%s/%s", data.Workload.WorkloadType, data.Workload.Name),
			SourceType:     "kubernetes",
			SourceInstance: fmt.Sprintf("cluster/%s", data.Cluster.Name),
			Time:           time,
			Data:           data,
		},
	}

	// Build ECST event
	ecstDiscoveryEvent := NewEcstBuilder().
		Header(header).
		Body(body).
		Build()
	return ecstDiscoveryEvent
}

func GenerateId(workspaceId string, configId string, data Data) string {
	scope := fmt.Sprintf(models.EventScopeFormat, workspaceId, configId)
	// workspace/{workspaceId}/configuration/{configurationId}/discoveryItem/service/kubernetes/workload/{clusterName}/{workloadType}/{workloadName}
	idString := fmt.Sprintf("%s/%s/%s/%s/%s/%s", scope, models.EventClassWorkload, data.Cluster.Name, data.Workload.WorkloadType, data.Workload.Name, data.NamespaceName)
	sum := sha256.Sum256([]byte(idString))
	id := hex.EncodeToString(sum[:])
	return id
}
