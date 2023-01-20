package models

type DiscoveryEvent struct {
	HeaderProperties HeaderProperties `json:"properties"`
	Body             DiscoveryBody    `json:"body"`
}

type DiscoveryBody struct {
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
	Class string `json:"class"`
	Type  string `json:"type"`
	Scope string `json:"scope"`
	Id    string `json:"id"`
}
