package iris

import (
	"fmt"
	"github.com/google/uuid"
	"time"
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
