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
	Cluster(cluster models.Cluster) EventBuilder
	TypeHeader(typeHeader string) EventBuilder
	Id(id string) EventBuilder
	Scope(scope string) EventBuilder
	Type(eventType string) EventBuilder
	Source(source string) EventBuilder
	Time(time string) EventBuilder
	Subject(subject string) EventBuilder

	Build() models.DiscoveryItem
}

type eventBuilder struct {
	c          models.Cluster
	id         string
	scope      string
	eventType  string
	source     string
	time       string
	subject    string
	typeHeader string
}

func New() EventBuilder {
	return &eventBuilder{}
}

func (eb *eventBuilder) TypeHeader(typeHeader string) EventBuilder {
	eb.typeHeader = typeHeader
	return eb
}

func (eb *eventBuilder) Cluster(cluster models.Cluster) EventBuilder {
	eb.c = cluster
	return eb
}

func (eb *eventBuilder) Id(id string) EventBuilder {
	eb.id = id
	return eb
}

func (eb *eventBuilder) Scope(scope string) EventBuilder {
	eb.scope = scope
	return eb
}

func (eb *eventBuilder) Type(eventType string) EventBuilder {
	eb.eventType = eventType
	return eb
}

func (eb *eventBuilder) Source(source string) EventBuilder {
	eb.source = source
	return eb
}

func (eb *eventBuilder) Time(time string) EventBuilder {
	eb.time = time
	return eb
}

func (eb *eventBuilder) Subject(subject string) EventBuilder {
	eb.subject = subject
	return eb
}

func (eb *eventBuilder) Build() models.DiscoveryItem {
	data := &models.Data{
		Cluster: eb.c,
	}
	header := &models.HeaderProperties{
		TypeHeader: eb.typeHeader,
	}
	return models.DiscoveryItem{
		ID:      eb.id,
		Scope:   eb.scope,
		Type:    eb.eventType,
		Source:  eb.source,
		Time:    eb.time,
		Subject: eb.subject,
		Data:    *data,
		Header:  *header,
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
