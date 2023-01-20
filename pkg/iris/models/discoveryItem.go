package models

type DiscoveryItem struct {
	ID      string        `json:"id"`
	Scope   string        `json:"scope"`
	Type    string        `json:"type"`
	Source  string        `json:"source"`
	Time    string        `json:"time"`
	Subject string        `json:"subject"`
	Data    DiscoveryData `json:"data"`
}

type DiscoveryData struct {
	Cluster Cluster `json:"cluster"`
}
