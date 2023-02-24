package models

type DiscoveryEvent struct {
	HeaderProperties HeaderProperties `json:"properties"`
	Body             DiscoveryBody    `json:"body"`
}

type DiscoveryBody struct {
	State State `json:"state"`
}

type State struct {
	Name           string `json:"name"`
	SourceType     string `json:"sourceType"`
	SourceInstance string `json:"sourceInstance"`
	Time           string `json:"time"`
	Data           Data   `json:"data"`
}

type Data struct {
	Cluster ClusterEcst `json:"cluster"`
}

type HeaderProperties struct {
	Class  string `json:"class"`
	Type   string `json:"type"`
	Scope  string `json:"scope"`
	Id     string `json:"id"`
	Action string `json:"action"`
}
