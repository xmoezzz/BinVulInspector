package task

import (
	"context"

	"bin-vul-inspector/pkg/models"
)

var (
	_ Handler = (*Bha)(nil) // bha server
)

type Handler interface {
	startJob(ctx context.Context, task *models.Task) error
	processResult(ctx context.Context, task *models.Task) error
}
