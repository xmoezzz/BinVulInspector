package jetstream

import (
	"github.com/nats-io/nats.go/jetstream"

	"bin-vul-inspector/pkg/nats"
)

type Client struct {
	jetstream.JetStream
}

func New(client *nats.Client) (c *Client, err error) {
	c = new(Client)
	c.JetStream, err = jetstream.New(client.Conn)
	return c, err
}
