package service

import (
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/client"
)

type Publisher struct {
	client client.Client
}

func newPublisher(c client.Client) *Publisher {
	return &Publisher{
		client: c,
	}
}

func (p *Publisher) New(topic string) micro.Event {
	return micro.NewEvent(topic, p.client)
}
