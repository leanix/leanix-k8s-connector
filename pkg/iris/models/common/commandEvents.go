package common

type CommandEvent struct {
	Properties CommandProperties `json:"properties"`
	Body       CommandBody       `json:"body"`
}

type CommandProperties struct {
	Type   string `json:"type"`
	Scope  string `json:"scope"`
	Action string `json:"action"`
}

type CommandBody struct{}
