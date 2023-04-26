package models

const (
	EVENT_TYPE_STATE    string = "state"
	EventTypeChange     string = "change"
	EventActionCreated  string = "created"
	EventActionUpdated  string = "updated"
	EventActionDeleted  string = "deleted"
	EventClassNamespace string = "discoveryItem/service/kubernetes/namespace"
	EventClassWorkload  string = "discoveryItem/service/kubernetes/workload"
	EventScopeFormat    string = "workspace/%s/configuration/%s"
)
