package models

type DiscoveryEvent struct {
	HeaderProperties HeaderProperties `json:"properties"`
	Body             DiscoveryItem    `json:"body"`
}

type DiscoveryItem struct {
	Source  string `json:"source"`
	Subject string `json:"subject"`
	Data    Data   `json:"data"`
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
