package models

type DiscoveryItem struct {
	ID      string           `json:"id"`
	Scope   string           `json:"scope"`
	Type    string           `json:"type"`
	Source  string           `json:"source"`
	Time    string           `json:"time"`
	Subject string           `json:"subject"`
	Data    Data             `json:"data"`
	Header  HeaderProperties `json:"header"`
}

type Data struct {
	Cluster Cluster `json:"cluster"`
}

type HeaderProperties struct {
	TypeHeader string `json:"type"`
}
