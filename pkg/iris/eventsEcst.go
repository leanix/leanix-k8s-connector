package iris

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
)

// ECST Discovery Items
type EcstEventBuilder interface {
	Header(header models.HeaderProperties) EcstEventBuilder
	Body(body models.DiscoveryBody) EcstEventBuilder

	Build() models.DiscoveryEvent
}

type CommandBuilder interface {
	Header(header models.CommandProperties) CommandBuilder
	Body(body models.CommandBody) CommandBuilder

	Build() models.CommandEvent
}

type ecstEventBuilder struct {
	header models.HeaderProperties
	body   models.DiscoveryBody
}

type commandBuilder struct {
	header models.CommandProperties
	body   models.CommandBody
}

// Discovery Event Builder
func NewEcstBuilder() EcstEventBuilder {
	return &ecstEventBuilder{}
}

func (eb *ecstEventBuilder) Header(header models.HeaderProperties) EcstEventBuilder {
	eb.header = header
	return eb
}

func (eb *ecstEventBuilder) Body(body models.DiscoveryBody) EcstEventBuilder {
	eb.body = body
	return eb
}

func (eb *ecstEventBuilder) Build() models.DiscoveryEvent {
	body := &models.DiscoveryBody{
		State: eb.body.State,
	}
	header := &models.HeaderProperties{
		Class: eb.header.Class,
		Type:  eb.header.Type,
		Scope: eb.header.Scope,
		Id:    eb.header.Id,
	}
	return models.DiscoveryEvent{
		HeaderProperties: *header,
		Body:             *body,
	}
}

//Command Event Builder
func NewCommand() CommandBuilder {
	return &commandBuilder{}
}

func (cb *commandBuilder) Header(header models.CommandProperties) CommandBuilder {
	cb.header = header
	return cb
}

func (cb *commandBuilder) Body(body models.CommandBody) CommandBuilder {
	cb.body = body
	return cb
}

func (cb *commandBuilder) Build() models.CommandEvent {
	header := &models.CommandProperties{
		Type:   cb.header.Type,
		Scope:  cb.header.Scope,
		Action: cb.header.Action,
	}
	body := &models.CommandBody{}

	return models.CommandEvent{
		Properties: *header,
		Body:       *body,
	}
}
