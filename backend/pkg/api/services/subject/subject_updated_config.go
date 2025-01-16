package subject

import (
	"context"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	stream "bin-vul-inspector/pkg/nats/jetstream"
)

type UpdatedConfig struct {
	js         *stream.Client
	streamName string
	subjects   []string

	once         sync.Once
	consumerName string
	consumer     jetstream.Consumer
}

func NewUpdatedConfig(client *stream.Client) *UpdatedConfig {
	return &UpdatedConfig{
		js:           client,
		streamName:   "config@jetstream",
		subjects:     []string{"jetstream.config.>"},
		consumerName: "consumer4config@jetstream",
	}
}

func (sub *UpdatedConfig) Name() string {
	return sub.streamName
}

func (sub *UpdatedConfig) ConsumerName() string {
	return sub.consumerName
}

func (sub *UpdatedConfig) Init(ctx context.Context) error {
	cfg := jetstream.StreamConfig{
		Name:      sub.streamName,
		Retention: jetstream.InterestPolicy,
		Subjects:  sub.subjects,
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	_, err := sub.js.CreateOrUpdateStream(ctxWithTimeout, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (sub *UpdatedConfig) Publish(ctx context.Context, msg Message) error {
	payload, err := msg.Payload()
	if err != nil {
		return err
	}

	subject := "jetstream.config.updated"
	_, err = sub.js.Publish(ctx, subject, payload)

	return err
}

func (sub *UpdatedConfig) createOrUpdateConsumer(ctx context.Context) (err error) {
	sub.once.Do(func() {
		cfg := jetstream.ConsumerConfig{
			Durable:       sub.consumerName,
			FilterSubject: "jetstream.config.>",
		}

		sub.consumer, err = sub.js.CreateOrUpdateConsumer(ctx, sub.streamName, cfg)
	})
	return err
}

func (sub *UpdatedConfig) FetchOne(ctx context.Context) (msg Message, ack func() error, err error) {
	err = sub.createOrUpdateConsumer(ctx)
	if err != nil {
		return nil, nil, err
	}

	res, err := sub.consumer.Fetch(1)
	if err != nil {
		return nil, nil, err
	}
	for rawMsg := range res.Messages() {
		if rawMsg != nil {
			msg = new(Config)
			if err = msg.Decode(rawMsg.Data()); err != nil {
				return nil, nil, err
			}
			ack = func() error {
				return rawMsg.Ack()
			}
			return msg, ack, nil
		}
		if res.Error() != nil {
			return nil, nil, res.Error()
		}
	}

	return nil, nil, nil
}

func (sub *UpdatedConfig) DeleteConsumer(ctx context.Context) error {
	return sub.js.DeleteConsumer(ctx, sub.streamName, sub.consumerName)
}
