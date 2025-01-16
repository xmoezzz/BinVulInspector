package nats

import (
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	*nats.Conn
}

func New(url string) (c *Client, err error) {
	c = new(Client)

	opts := []nats.Option{
		nats.Timeout(5 * time.Second),
	}
	c.Conn, err = nats.Connect(url, opts...)
	return c, err
}

func (s *Client) Close() error {
	return s.Conn.Drain()
}
