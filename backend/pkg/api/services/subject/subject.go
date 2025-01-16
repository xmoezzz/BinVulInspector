package subject

import (
	"context"
)

type Subject interface {
	Name() string
	Init(ctx context.Context) error
	Publish(context.Context, Message) error
	FetchOne(context.Context) (Message, func() error, error)
}
