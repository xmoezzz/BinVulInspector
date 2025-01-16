package kit

import (
	"context"
	"sync"

	"github.com/minio/minio-go/v7"

	"bin-vul-inspector/pkg/config"
	"bin-vul-inspector/pkg/log"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/nats"
	"bin-vul-inspector/pkg/nats/jetstream"
)

type Kit struct {
	Logger    log.Logger
	Config    *config.App
	Mongo     *mongo.Client
	Minio     *minio.Client
	Nats      *nats.Client
	JetStream *jetstream.Client
}

func (kit *Kit) Close(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := kit.Mongo.Close(ctx); err != nil {
			kit.Logger.Errorf("close mongo client error: %w", err)
		}
	}()

	wg.Wait()
}
