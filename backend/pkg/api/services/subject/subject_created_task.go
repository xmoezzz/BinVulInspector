package subject

import (
	"context"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	stream "bin-vul-inspector/pkg/nats/jetstream"
)

type CreatedTask struct {
	js         *stream.Client
	streamName string
	subjects   []string

	once         sync.Once
	consumerName string
	consumer     jetstream.Consumer
}

func NewCreatedTask(client *stream.Client) *CreatedTask {
	return &CreatedTask{
		js:           client,
		streamName:   "tasks@jetstream",
		subjects:     []string{"jetstream.tasks.>"},
		consumerName: "consumer4tasks@jetstream",
	}
}

func (sub *CreatedTask) Name() string {
	return sub.streamName
}

func (sub *CreatedTask) ConsumerName() string {
	return sub.consumerName
}

func (sub *CreatedTask) Init(ctx context.Context) error {
	cfg := jetstream.StreamConfig{
		Name:      sub.streamName,
		Retention: jetstream.WorkQueuePolicy,
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

func (sub *CreatedTask) Publish(ctx context.Context, msg Message) error {
	payload, err := msg.Payload()
	if err != nil {
		return err
	}

	subject := "jetstream.tasks.created"
	_, err = sub.js.Publish(ctx, subject, payload)

	return err
}

func (sub *CreatedTask) createOrUpdateConsumer(ctx context.Context) (err error) {
	sub.once.Do(func() {
		cfg := jetstream.ConsumerConfig{
			Durable:       sub.consumerName,
			FilterSubject: "jetstream.tasks.>",
		}

		sub.consumer, err = sub.js.CreateOrUpdateConsumer(ctx, sub.streamName, cfg)
	})
	return err
}

func (sub *CreatedTask) FetchOne(ctx context.Context) (msg Message, ack func() error, err error) {
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
			msg = new(Task)
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

func (sub *CreatedTask) DeleteConsumer(ctx context.Context) error {
	return sub.js.DeleteConsumer(ctx, sub.streamName, sub.consumerName)
}
