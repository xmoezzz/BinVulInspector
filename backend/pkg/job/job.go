package job

import (
	"context"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/job/task"
)

type Job interface {
	Name() string
	Run(context.Context)
	Close()
	Trigger(string)
}

type Jobs []Job

func NewJobs(kit *kit.Kit) Jobs {
	return []Job{
		task.NewJob(kit),
	}
}

func (jobs Jobs) Run(ctx context.Context) {
	for i := range jobs {
		jobs[i].Trigger("start to running")
		jobs[i].Run(ctx)
	}
}

func (jobs Jobs) Close() {
	for i := range jobs {
		jobs[i].Trigger("stop running")
		jobs[i].Close()
		jobs[i].Trigger("stopped")
	}
}
