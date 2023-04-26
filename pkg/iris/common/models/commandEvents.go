package models

import "fmt"

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

type CommandBuilder interface {
	Header(header CommandProperties) CommandBuilder
	Body(body CommandBody) CommandBuilder

	Build() CommandEvent
}

type commandBuilder struct {
	header CommandProperties
	body   CommandBody
}

// Command Event Builder
func NewCommand() CommandBuilder {
	return &commandBuilder{}
}

func (cb *commandBuilder) Header(header CommandProperties) CommandBuilder {
	cb.header = header
	return cb
}

func (cb *commandBuilder) Body(body CommandBody) CommandBuilder {
	cb.body = body
	return cb
}

func (cb *commandBuilder) Build() CommandEvent {
	header := &CommandProperties{
		Type:   cb.header.Type,
		Scope:  cb.header.Scope,
		Action: cb.header.Action,
	}
	body := &CommandBody{}

	return CommandEvent{
		Properties: *header,
		Body:       *body,
	}
}

// Command Events
func CreateStartReplay(workspaceId string, config KubernetesConfig) CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("startReplay")
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, config.ID)
	header := CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	startReplayEvent := NewCommand().Header(header).Build()
	return startReplayEvent
}

func CreateEndReplay(workspaceId string, config KubernetesConfig) CommandEvent {
	// Metadata for the command event
	eventType := fmt.Sprintf("command")
	action := fmt.Sprintf("endReplay")
	scope := fmt.Sprintf(EventScopeFormat, workspaceId, config.ID)
	header := CommandProperties{
		Type:   eventType,
		Action: action,
		Scope:  scope,
	}

	endReplayEvent := NewCommand().Header(header).Build()
	return endReplayEvent
}
