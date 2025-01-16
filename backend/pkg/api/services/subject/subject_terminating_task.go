package subject

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	localNats "bin-vul-inspector/pkg/nats"
)

type TerminatingTask struct {
	nats        *localNats.Client
	subjectName string

	once         sync.Once
	subscription *nats.Subscription
}

func NewTerminatingTask(client *localNats.Client) *TerminatingTask {
	return &TerminatingTask{
		nats:        client,
		subjectName: "tasks.terminating",
	}
}

func (sub *TerminatingTask) SubjectName() string {
	return sub.subjectName
}

func (sub *TerminatingTask) Request(msg Message) ([]byte, error) {
	payload, err := msg.Payload()
	if err != nil {
		return nil, err
	}

	rawMsg, err := sub.nats.Request(sub.subjectName, payload, 3*time.Minute)
	if err != nil {
		if errors.Is(err, nats.ErrNoResponders) {
			return nil, nil
		}
		return nil, err
	}
	return rawMsg.Data, err
}

func (sub *TerminatingTask) Subscribe(f func(*Task) bool) (err error) {
	sub.once.Do(func() {
		sub.subscription, err = sub.nats.Subscribe(sub.subjectName, func(rawMsg *nats.Msg) {
			msg := new(Task)
			_ = msg.Decode(rawMsg.Data)
			exist := f(msg)
			_ = rawMsg.Respond([]byte(strconv.FormatBool(exist)))
		})
	})
	return err
}

func (sub *TerminatingTask) Unsubscribe() error {
	return sub.subscription.Unsubscribe()
}
