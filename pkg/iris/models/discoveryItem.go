package models

type DiscoveryEvent struct {
	HeaderProperties HeaderProperties `json:"properties"`
	Body             DiscoveryItem    `json:"body"`
}

type DiscoveryItem struct {
	State State `json:"state"`
	Data  Data  `json:"data"`
}

type State struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

type Data struct {
	Cluster Cluster `json:"cluster"`
}

type HeaderProperties struct {
	HeaderClass string `json:"class"`
	HeaderType  string `json:"type"`
	HeaderScope string `json:"scope"`
	HeaderId    string `json:"id"`
}
