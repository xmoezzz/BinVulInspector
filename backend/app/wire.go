//go:build wireinject

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/wire"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/api"
	"bin-vul-inspector/pkg/config"
	"bin-vul-inspector/pkg/log"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/nats"
	"bin-vul-inspector/pkg/nats/jetstream"
)

func provideLogger() (log.Logger, error) {
	logger, err := log.NewConsoleLogger("debug")
	if err != nil {
		return nil, err
	}

	logger.Info(provideSuccess("logger"))
	return logger, nil
}

func provideError(component string, err error) error {
	return fmt.Errorf("provide %10s error %w", component, err)
}

func provideSuccess(component string) error {
	return fmt.Errorf("provide %10s success", component)
}

func provideConfig(configFilepath string, logger log.Logger) (*config.App, error) {
	component := fmt.Sprintf("config(%s)", configFilepath)

	conf, err := config.LoadApp(configFilepath)
	if err != nil {
		return nil, provideError(component, err)
	}
	logger.Info(provideSuccess(component))
	logger.Debugf("loaded %s: %+v", component, conf)
	return conf, nil
}

func provideMongo(ctx context.Context, conf *config.App, logger log.Logger) (*mongo.Client, error) {
	component := "mongo"
	for {
		client, err := mongo.NewClient(ctx, conf.MongoDB.URI, logger)
		if err != nil {
			logger.Error(provideError(component, err))

			time.Sleep(10 * time.Second)
			continue
		}
		logger.Info(provideSuccess(component))
		return client, nil
	}
}

func provideMinio(conf *config.App, logger log.Logger) (*minio.Client, error) {
	component := "minio"

	client, err := minio.New(
		conf.Minio.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(conf.Minio.AccessKeyID, conf.Minio.SecretAccessKey, ""),
			Secure: conf.Minio.UseSSL,
		},
	)
	if err != nil {
		return nil, provideError(component, err)
	}
	logger.Info(provideSuccess(component))
	return client, nil
}

func provideNats(conf *config.App, logger log.Logger) (*nats.Client, error) {
	component := "nats"

	stream, err := nats.New(conf.Nats.Url)
	if err != nil {
		return nil, provideError(component, err)
	}
	logger.Info(provideSuccess(component))
	return stream, nil
}

func provideJetStream(client *nats.Client, logger log.Logger) (*jetstream.Client, error) {
	component := "jetstream"

	stream, err := jetstream.New(client)
	if err != nil {
		return nil, provideError(component, err)
	}
	logger.Info(provideSuccess(component))
	return stream, nil
}

func NewKit(ctx context.Context, configFilepath string) (*kit.Kit, error) {
	wire.Build(
		provideLogger,
		provideConfig,
		provideMongo,
		provideMinio,
		provideNats,
		provideJetStream,

		wire.Struct(
			new(kit.Kit),
			"Logger",
			"Config",
			"Mongo",
			"MinIO",
			"Nats",
			"JetStream",
		),
	)
	return nil, nil
}
func NewApp(kit *kit.Kit) (*App, error) {
	wire.Build(
		api.NewHttpHandler,
		New,
	)
	return nil, nil
}
