package workload

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/common"
	time2 "time"
)

const (
	EVENT_TYPE_STATE   string = "state"
	EventTypeChange    string = "change"
	EventActionCreated string = "created"
	EventActionUpdated string = "updated"
	EventActionDeleted string = "deleted"
	EventClassWorkload string = "discoveryItem/service/kubernetes/workload"
	EventScopeFormat   string = "workspace/%s/configuration/%s"
)

// ECST Discovery Items
type EcstEventBuilder interface {
	Header(header HeaderProperties) EcstEventBuilder
	Body(body DiscoveryBody) EcstEventBuilder

	Build() DiscoveryEvent
}

type CommandBuilder interface {
	Header(header common.CommandProperties) CommandBuilder
	Body(body common.CommandBody) CommandBuilder

	Build() common.CommandEvent
}

type ecstEventBuilder struct {
	header HeaderProperties
	body   DiscoveryBody
}

type EcstWorkloadEventBuilder interface {
	Header(header HeaderProperties) EcstEventBuilder
	Body(body DiscoveryBody) EcstEventBuilder

	Build() DiscoveryEvent
}

type commandBuilder struct {
	header common.CommandProperties
	body   common.CommandBody
}

// NewEcstBuilder Discovery Event Builder
func NewEcstBuilder() EcstEventBuilder {
	return &ecstEventBuilder{}
}

func (eb *ecstEventBuilder) Header(header HeaderProperties) EcstEventBuilder {
	eb.header = header
	return eb
}

func (eb *ecstEventBuilder) Body(body DiscoveryBody) EcstEventBuilder {
	eb.body = body
	return eb
}

func (eb *ecstEventBuilder) Build() DiscoveryEvent {
	body := &DiscoveryBody{
		State: eb.body.State,
	}

	header := &HeaderProperties{
		Class:  eb.header.Class,
		Type:   eb.header.Type,
		Scope:  eb.header.Scope,
		Id:     eb.header.Id,
		Action: eb.header.Action,
	}
	return DiscoveryEvent{
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

// CreateEcstWorkloadDiscoveryEvent ECST Discovery Items
func CreateEcstWorkloadDiscoveryEvent(eventType string, changeAction string, data Data, runId string, workspaceId string, configId string) DiscoveryEvent {
	// Metadata for the event
	id := GenerateWorkloadId(workspaceId, configId, data)
	time := time2.Now().Format(time2.RFC3339)

	var header HeaderProperties
	if eventType == EventTypeChange {
		header = HeaderProperties{
			Id:     id,
			Scope:  fmt.Sprintf(EventScopeFormat, workspaceId, configId),
			Class:  EventClassWorkload,
			Type:   eventType,
			Action: changeAction,
		}
	} else {
		header = HeaderProperties{
			Id:    id,
			Scope: fmt.Sprintf(EventScopeFormat, workspaceId, configId),
			Class: EventClassWorkload,
			Type:  eventType,
		}
	}

	body := DiscoveryBody{
		State: State{
			Name:           data.Workload[0].ClusterName,
			SourceType:     "kubernetes",
			SourceInstance: fmt.Sprintf("cluster/%s", data.Workload[0].ClusterName),
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

func GenerateWorkloadId(workspaceId string, configId string, data Data) string {
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, configId)
	// workspace/{workspaceId}/configuration/{configurationId}/discoveryItem/service/kubernetes/{clusterName}/{namespaceName}
	idString := fmt.Sprintf("%s/%s/%s/%s", scope, EventClassWorkload, data.Workload[0].ClusterName, data.Workload[0].WorkloadName)
	sum := sha256.Sum256([]byte(idString))
	id := hex.EncodeToString(sum[:])
	return id
}
