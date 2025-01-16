package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"bin-vul-inspector/pkg/log"
)

const (
	configsCollection = "configs"

	tasksCollection = "tasks"

	bhaFuncsCollection       = "bha_funcs"
	bhaFuncResultsCollection = "bha_func_results"
	bhaModelsCollection      = "bha_models"
)

type Client struct {
	client     *mongo.Client
	AuthSource string
}

func NewClient(ctx context.Context, uri string, logger log.Logger) (*Client, error) {
	opts := options.Client()
	opts.ApplyURI(uri)
	opts.SetServerSelectionTimeout(2 * time.Second)
	if logger != nil {
		opts.SetMonitor(&event.CommandMonitor{
			Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
				// logger.Info(evt.Command.String())
			},
			Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
				// logger.Info(evt.Reply.String())
			},
			Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
				logger.Info(evt.Failure)
			},
		})
	}

	client, err := mongo.Connect(ctx, opts)
	if client == nil || err != nil {
		if err != nil {
			return nil, fmt.Errorf("connect to %s failed: %v", uri, err)
		}
		return nil, fmt.Errorf("connect to %s failed", uri)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("mongo ping error, %w", err)
	}

	var authSource string
	if opts.Auth != nil {
		authSource = opts.Auth.AuthSource
	}

	return &Client{
		client:     client,
		AuthSource: authSource,
	}, nil
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
