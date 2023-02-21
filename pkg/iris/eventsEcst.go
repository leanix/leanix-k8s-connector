package iris

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	time2 "time"
)

const (
	EVENT_TYPE_STATE   string = "state"
	EventTypeChange    string = "change"
	EventActionCreated string = "created"
	EventActionUpdated string = "updated"
	EventActionDeleted string = "deleted"
	EventClass         string = "discoveryItem/service/kubernetes"
	EventScopeFormat   string = "workspace/%s/configuration/%s"
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

type commandBuilder struct {
	header models.CommandProperties
	body   models.CommandBody
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
func CreateEcstDiscoveryEvent(eventType string, changeAction string, data models.Data, runId string, workspaceId string, configId string) models.DiscoveryEvent {
	// Metadata for the event
	id := GenerateId(workspaceId, configId, data)
	time := time2.Now().Format(time2.RFC3339)

	var header models.HeaderProperties
	if eventType == EventTypeChange {
		header = models.HeaderProperties{
			Id:     id,
			Scope:  fmt.Sprintf(EventScopeFormat, workspaceId, configId),
			Class:  EventClass,
			Type:   eventType,
			Action: changeAction,
		}
	} else {
		header = models.HeaderProperties{
			Id:    id,
			Scope: fmt.Sprintf(EventScopeFormat, workspaceId, configId),
			Class: EventClass,
			Type:  eventType,
		}
	}

	body := models.DiscoveryBody{
		State: models.State{
			Name:   data.Cluster.Namespace,
			Source: fmt.Sprintf("kubernetes/%s#%s/", data.Cluster.Name, runId),
			Time:   time,
			Data:   data,
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
	idString := fmt.Sprintf("%s/%s/%s/%s", scope, EventClass, data.Cluster.Name, data.Cluster.Namespace)
	sum := sha256.Sum256([]byte(idString))
	id := hex.EncodeToString(sum[:])
	return id
}

// Command Events
func CreateStartReplay(workspaceId string, config kubernetesConfig) models.CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("startReplay")
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, config.ID)
	header := models.CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	startReplayEvent := NewCommand().Header(header).Build()
	return startReplayEvent
}

func CreateEndReplay(workspaceId string, config kubernetesConfig) models.CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("endReplay")
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, config.ID)
	header := models.CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	endReplayEvent := NewCommand().Header(header).Build()
	return endReplayEvent
}

func CreateDiscoveryItem(cluster models.Cluster, source string, scope string) models.DiscoveryItem {
	// Metadata for the event
	id := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s", cluster.Name, cluster.Namespace.Name))))
	subject := fmt.Sprintf("namespace/%s", cluster.Namespace.Name)
	time := time2.Now().Format(time2.RFC3339)

	// Build service/softwareArtifact event
	discoveryEvent := New().
		Id(id).
		Source(source).
		Subject(subject).
		Type(typeAsK8sNamespace).
		Scope(scope).
		Time(time).
		Cluster(cluster).
		Build()
	return discoveryEvent
}
