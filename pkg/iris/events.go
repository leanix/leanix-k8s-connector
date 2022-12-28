package iris

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
)

const (
	typeAsK8sNamespace string = "leanix.vsm.item-discovered.kubernetes.namespace"
)

// struct to extend Log with RunId
type Log struct {
	// Root        *Logger
	// Output      *Output
	RunId       string
	WorkspaceId string
}

func GenerateRunId() string {
	return uuid.New().String()
}

type EventBuilder interface {
	Header(header models.HeaderProperties) EventBuilder
	Body(body models.DiscoveryItem) EventBuilder

	Build() models.DiscoveryEvent
}

type CommandBuilder interface {
	Header(header models.CommandProperties) CommandBuilder
	Body(body models.CommandBody) CommandBuilder

	Build() models.CommandEvent
}

type eventBuilder struct {
	header models.HeaderProperties
	body   models.DiscoveryItem
}

type commandBuilder struct {
	header models.CommandProperties
	body   models.CommandBody
}

// Discovery Event Builder
func New() EventBuilder {
	return &eventBuilder{}
}

func (eb *eventBuilder) Header(header models.HeaderProperties) EventBuilder {
	eb.header = header
	return eb
}

func (eb *eventBuilder) Body(body models.DiscoveryItem) EventBuilder {
	eb.body = body
	return eb
}

func (eb *eventBuilder) Build() models.DiscoveryEvent {
	body := &models.DiscoveryItem{
		State: eb.body.State,
	}
	header := &models.HeaderProperties{
		Class: eb.header.Class,
		Type:  eb.header.Type,
		Scope: eb.header.Scope,
		Id:    eb.header.Id,
	}
	return models.DiscoveryEvent{
		HeaderProperties: *header,
		Body:             *body,
	}
}

//Command Event Builder
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

type StatusItem struct {
	ID              string      `json:"id"`
	Scope           string      `json:"scope"`
	Type            string      `json:"type"`
	Source          string      `json:"source"`
	Time            string      `json:"time"`
	DataContentType string      `json:"datacontenttype"`
	DataSchema      string      `json:"dataschema"`
	Subject         string      `json:"subject"`
	Data            interface{} `json:"data"`
}

func NewStatusEvent(configurationId string, runId string, workspaceId string, runstatus string, message string) *StatusItem {

	Id := fmt.Sprintf("%s", configurationId)
	Subject := fmt.Sprintf(runstatus)
	Type := fmt.Sprintf("leanix.vsm.item-logged.status")
	Scope := fmt.Sprintf("workspace/%s", workspaceId)
	Source := fmt.Sprintf("kubernetes/%s#%s", configurationId, runId)
	Time := fmt.Sprintf(time.Now().Format(time.RFC3339))
	DataContentType := fmt.Sprintf("application/json")
	DataSchema := fmt.Sprintf("/vsm-iris/schemas/feedback-items/v1")

	var StatusData = make(map[string]interface{})
	StatusData["status"] = runstatus
	StatusData["message"] = message

	return &StatusItem{
		ID:              Id,
		Scope:           Scope,
		Type:            Type,
		Source:          Source,
		Time:            Time,
		Subject:         Subject,
		DataContentType: DataContentType,
		DataSchema:      DataSchema,
		Data:            StatusData,
	}
}

func NewAdminLogEvent(configurationId string, runId string, workspaceId string, loglevel string, message string) *StatusItem {

	Id := fmt.Sprintf("%s", configurationId)
	Subject := fmt.Sprintf(loglevel)
	Type := fmt.Sprintf("leanix.vsm.item-logged.admin")
	Scope := fmt.Sprintf("workspace/%s", workspaceId)
	Source := fmt.Sprintf("kubernetes/%s#%s", configurationId, runId)
	Time := fmt.Sprintf(time.Now().Format(time.RFC3339))
	DataContentType := fmt.Sprintf("application/json")
	DataSchema := fmt.Sprintf("/vsm-iris/schemas/feedback-items/v1")

	var StatusData = make(map[string]interface{})
	StatusData["level"] = loglevel
	StatusData["message"] = message

	return &StatusItem{
		ID:              Id,
		Scope:           Scope,
		Type:            Type,
		Source:          Source,
		Time:            Time,
		Subject:         Subject,
		DataContentType: DataContentType,
		DataSchema:      DataSchema,
		Data:            StatusData,
	}
}
