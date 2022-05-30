package iris

type DiscoveryItem struct {
	ID      string      `json:"id"`
	Scope   string      `json:"scope"`
	Type    string      `json:"type"`
	Source  string      `json:"source"`
	Time    string      `json:"time"`
	Subject string      `json:"subject"`
	Data    interface{} `json:"data"`
}

func GenerateDiscoveryItem(id string, scope string, t string, source string, time string, subject string, data interface{}) *DiscoveryItem {
	return &DiscoveryItem{
		ID:      id,
		Scope:   scope,
		Type:    t,
		Source:  source,
		Time:    time,
		Subject: subject,
		Data:    data,
	}
}
