package main

import (
	"context"
	"fmt"

	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"bin-vul-inspector/app"
	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/api/services"
	"bin-vul-inspector/pkg/api/services/subject"
	"bin-vul-inspector/pkg/job"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/mongo"
)

//go:generate swag fmt --exclude scs-workdir
//go:generate swag init --generalInfo=server.go --output=../../docs --dir=.,../../pkg/api/v1,../../pkg/models

//	@title						bin-vul-inspector Server API
//	@version					1.0
//	@description				This is an bin-vul-inspector server.
//	@BasePath					/scs/api/v1
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization

func New() *cobra.Command {
	opts := new(options)
	cmd := &cobra.Command{
		Use:          "server",
		Short:        "bin-vul-inspector Server",
		Long:         "bin-vul-inspector Server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run(context.Background())
		},
	}

	cmd.Flags().StringVarP(&opts.config, "config", "c", "config.yaml", "config file")
	cmd.Flags().BoolVar(&opts.server, "server", true, "enable server")
	cmd.Flags().BoolVar(&opts.runner, "runner", true, "enable runner")

	return cmd
}

type options struct {
	config string
	server bool
	runner bool
}

func (opts *options) run(ctx context.Context) (err error) {
	var toolkit *kit.Kit
	if toolkit, err = app.NewKit(ctx, opts.config); err != nil {
		return err
	}
	defer toolkit.Close(ctx)

	// 初始化系统
	if err = initApp(ctx, toolkit); err != nil {
		return err
	}

	// server mode
	if opts.server {
		var server *app.App
		if server, err = app.NewApp(toolkit); err != nil {
			return err
		}
		defer server.Close(ctx)
		go server.Serve()
	}

	// runner mode
	if opts.runner {
		jobs := job.NewJobs(toolkit)
		jobs.Run(ctx)
		defer jobs.Close()
	}

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return nil
}

func initApp(ctx context.Context, kit *kit.Kit) (err error) {
	// 初始化mongo数据结构
	if err = mongo.InitCollections(ctx, kit.Mongo); err != nil {
		return fmt.Errorf("init mongo indices error: %w", err)
	}

	// 固化系统配置或者加载系统配置
	if err = services.NewConfig(kit).InitOrLoad(ctx, kit.Config); err != nil {
		return fmt.Errorf("init or load config error: %w", err)
	}

	if err = minio.InitBucket(ctx, kit.Minio); err != nil {
		return fmt.Errorf("init minio bucket error: %w", err)
	}

	if err = services.NewBha(kit).Init(ctx); err != nil {
		return fmt.Errorf("init bha model error: %w", err)
	}

	streams := []subject.Subject{
		subject.NewCreatedTask(kit.JetStream),
		subject.NewUpdatedConfig(kit.JetStream),
	}
	for _, stream := range streams {
		if err = stream.Init(ctx); err != nil {
			return fmt.Errorf("init nats jetStream %s error: %w", stream.Name(), err)
		} else {
			kit.Logger.Infof("init nats jetStream %s success", stream.Name())
		}
	}

	return nil
}
