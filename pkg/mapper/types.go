package mapper

// KubernetesObject describes an object that is read from the Kubernetes api
type KubernetesObject struct {
	Type string      `json:"type,omitempty"`
	ID   string      `json:"id,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// LDIF (LEAN Data Interchange Format) represents the output file generated by the connector
type LDIF struct {
	ConnectorID      string             `json:"connectorId,omitempty"`
	ConnectorType    string             `json:"connectorType,omitempty"`
	ConnectorVersion string             `json:"connectorVersion,omitempty"`
	LxVersion        string             `json:"lxVersion,omitempty"`
	LxWorkspace      string             `json:"lxWorkspace,omitempty"`
	Description      string             `json:"description,omitempty"`
	Content          []KubernetesObject `json:"content,omitempty"`
}
