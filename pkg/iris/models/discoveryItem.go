package models

type DiscoveryEvent struct {
	HeaderProperties HeaderProperties `json:"properties"`
	Body             DiscoveryItem    `json:"body"`
}

type DiscoveryItem struct {
	State State `json:"state"`
}

type State struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Time   string `json:"time"`
	Data   Data   `json:"data"`
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
