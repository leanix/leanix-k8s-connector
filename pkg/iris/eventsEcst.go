package iris

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/common"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/namespace"
	time2 "time"
)

const (
	EVENT_TYPE_STATE    string = "state"
	EventTypeChange     string = "change"
	EventActionCreated  string = "created"
	EventActionUpdated  string = "updated"
	EventActionDeleted  string = "deleted"
	EventClassNamespace string = "discoveryItem/service/kubernetes/namespace"
	EventScopeFormat    string = "workspace/%s/configuration/%s"
)

// ECST Discovery Items
type EcstEventBuilder interface {
	Header(header models.HeaderProperties) EcstEventBuilder
	Body(body models.DiscoveryBody) EcstEventBuilder

	Build() models.DiscoveryEvent
}

type CommandBuilder interface {
	Header(header common.CommandProperties) CommandBuilder
	Body(body common.CommandBody) CommandBuilder

	Build() common.CommandEvent
}

type ecstEventBuilder struct {
	header models.HeaderProperties
	body   models.DiscoveryBody
}

type commandBuilder struct {
	header common.CommandProperties
	body   common.CommandBody
}

// Discovery Event Builder
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

func (cb *commandBuilder) Header(header common.CommandProperties) CommandBuilder {
	cb.header = header
	return cb
}

func (cb *commandBuilder) Body(body common.CommandBody) CommandBuilder {
	cb.body = body
	return cb
}

func (cb *commandBuilder) Build() common.CommandEvent {
	header := &common.CommandProperties{
		Type:   cb.header.Type,
		Scope:  cb.header.Scope,
		Action: cb.header.Action,
	}
	body := &common.CommandBody{}

	return common.CommandEvent{
		Properties: *header,
		Body:       *body,
	}
}

// CreateEcstDiscoveryEvent ECST Discovery Items
func CreateEcstDiscoveryEvent(eventType string, changeAction string, data models.Data, workspaceId string, configId string) models.DiscoveryEvent {
	// Metadata for the event
	id := GenerateId(workspaceId, configId, data)
	time := time2.Now().Format(time2.RFC3339)

	var header models.HeaderProperties
	if eventType == EventTypeChange {
		header = models.HeaderProperties{
			Id:     id,
			Scope:  fmt.Sprintf(EventScopeFormat, workspaceId, configId),
			Class:  EventClassNamespace,
			Type:   eventType,
			Action: changeAction,
		}
	} else {
		header = models.HeaderProperties{
			Id:    id,
			Scope: fmt.Sprintf(EventScopeFormat, workspaceId, configId),
			Class: EventClassNamespace,
			Type:  eventType,
		}
	}

	body := models.DiscoveryBody{
		State: models.State{
			Name:           data.Cluster.Namespace,
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

func GenerateId(workspaceId string, configId string, data models.Data) string {
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, configId)
	// workspace/{workspaceId}/configuration/{configurationId}/discoveryItem/service/kubernetes/{clusterName}/{namespaceName}
	idString := fmt.Sprintf("%s/%s/%s/%s", scope, EventClassNamespace, data.Cluster.Name, data.Cluster.Namespace)
	sum := sha256.Sum256([]byte(idString))
	id := hex.EncodeToString(sum[:])
	return id
}

// Command Events
func CreateStartReplay(workspaceId string, config kubernetesConfig) common.CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("startReplay")
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, config.ID)
	header := common.CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	startReplayEvent := NewCommand().Header(header).Build()
	return startReplayEvent
}

func CreateEndReplay(workspaceId string, config kubernetesConfig) common.CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("endReplay")
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, config.ID)
	header := common.CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	endReplayEvent := NewCommand().Header(header).Build()
	return endReplayEvent
}
