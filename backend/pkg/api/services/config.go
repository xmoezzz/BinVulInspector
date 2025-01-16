package services

import (
	"context"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/api/services/subject"
	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/config"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
)

type Config struct {
	*kit.Kit
}

func NewConfig(kit *kit.Kit) *Config {
	return &Config{
		Kit: kit,
	}
}

func (svc *Config) InitOrLoad(ctx context.Context, conf *config.App) error {
	m, err := mongo.NewConfig(svc.Mongo).Latest(ctx)
	if err != nil {
		return err
	}

	var changed bool
	if m == nil {
		m = new(models.Config)
		m.Task.Concurrent = conf.Task.Concurrent
		changed = true
	} else {
		conf.Task.Concurrent = m.Task.Concurrent
	}
	if m.Task.ScaTimeout != nil {
		conf.Task.ScaTimeout = utils.Sec2Duration(*m.Task.ScaTimeout)
	} else {
		m.Task.ScaTimeout = pointer.Of(utils.Duration2Sec(conf.Task.ScaTimeout))
		changed = true
	}
	if m.Task.SastTimeout != nil {
		conf.Task.SastTimeout = utils.Sec2Duration(*m.Task.SastTimeout)
	} else {
		m.Task.SastTimeout = pointer.Of(utils.Duration2Sec(conf.Task.SastTimeout))
		changed = true
	}
	if m.Task.BhaTimeout != nil {
		conf.Task.BhaTimeout = utils.Sec2Duration(*m.Task.BhaTimeout)
	} else {
		m.Task.BhaTimeout = pointer.Of(utils.Duration2Sec(conf.Task.BhaTimeout))
		changed = true
	}

	if !changed {
		return nil
	}

	_, err = mongo.NewConfig(svc.Mongo).Insert(ctx, m)
	return err
}

func (svc *Config) Update(ctx context.Context, req *dto.UpdateSettingsRequest) error {
	m, err := mongo.NewConfig(svc.Mongo).Latest(ctx)
	if err != nil {
		return err
	}

	var changed bool
	if req.Concurrent != nil && *req.Concurrent != m.Task.Concurrent {
		changed = true
		m.Task.Concurrent = *req.Concurrent
	}
	if req.ScaTimeout != nil && *req.ScaTimeout != *m.Task.ScaTimeout {
		changed = true
		m.Task.ScaTimeout = pointer.Of(*req.ScaTimeout)
	}
	if req.SastTimeout != nil && *req.SastTimeout != *m.Task.SastTimeout {
		changed = true
		m.Task.SastTimeout = pointer.Of(*req.SastTimeout)
	}
	if req.BhaTimeout != nil && *req.BhaTimeout != *m.Task.BhaTimeout {
		changed = true
		m.Task.BhaTimeout = pointer.Of(*req.BhaTimeout)
	}

	if !changed {
		return nil
	}
	_, err = mongo.NewConfig(svc.Mongo).Insert(ctx, m)
	if err != nil {
		return err
	}

	err = subject.NewUpdatedConfig(svc.JetStream).Publish(ctx, subject.NewConfig(m))
	if err != nil {
		svc.Logger.Errorf("publish config update error, %w", err)
	} else {
		svc.Logger.Debugf("publish config update success")
	}

	return nil
}
