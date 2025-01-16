package task

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/api/services/subject"
	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
)

type Job struct {
	*kit.Kit
	name               string
	closeEvent         chan struct{}
	wg                 sync.WaitGroup
	subject            *subject.CreatedTask
	configSubject      *subject.UpdatedConfig
	terminatingSubject *subject.TerminatingTask
	ctxCache           sync.Map
	limiters           Limiters
	pool               *Limiter
	handler            map[string]Handler
}

func NewJob(kit *kit.Kit) *Job {
	job := &Job{
		Kit:                kit,
		name:               "tasks@job",
		closeEvent:         make(chan struct{}),
		subject:            subject.NewCreatedTask(kit.JetStream),
		configSubject:      subject.NewUpdatedConfig(kit.JetStream),
		terminatingSubject: subject.NewTerminatingTask(kit.Nats),
		handler: map[string]Handler{
			constant.TypeBha: NewBHA(kit),
		},
	}

	job.pool = NewLimiter(kit.Config.Task.Concurrent)
	return job
}

func (job *Job) Name() string {
	return job.name
}

func (job *Job) Trigger(event string) {
	job.Logger.Infof("%15s job: %s.", job.Name(), event)
}

func (job *Job) Run(ctx context.Context) {
	job.wg.Add(1)
	go func(ctx context.Context) {
		defer job.wg.Done()

		ctxWithCancel, cancel := context.WithCancel(ctx)
		job.runConfigConsumer(ctxWithCancel)
		job.runTasksConsumer(ctxWithCancel)
		job.runTerminating()

		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-job.closeEvent:
				cancel()
				job.debugf("job context canceled")
				return
			case <-ticker.C:
				job.processTimeoutTasks(ctxWithCancel)
				job.processDeletedModel(ctxWithCancel)
			}
		}
	}(ctx)
}

func (job *Job) Close() {
	close(job.closeEvent)
	job.wg.Wait()
	_ = job.subject.DeleteConsumer(context.Background())
	_ = job.configSubject.DeleteConsumer(context.Background())
	_ = job.terminatingSubject.Unsubscribe()
	job.terminateProcessTasks()
}

func (job *Job) runTasksConsumer(ctx context.Context) {

	job.wg.Add(1)
	go func(ctx context.Context) {
		defer job.wg.Done()

		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				job.debugf("tasks consumer context canceled")
				return
			case <-ticker.C:
				job.debugf("tasks consumer has been running. job.capacity: %d, job.running: %d", job.pool.capacity, job.pool.running)
			default:
			}
			job.tasksConsumer(ctx)
		}
	}(ctx)
}

func (job *Job) tasksConsumer(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			job.errorf("tasks consumer panic: %v\n%s", e, debug.Stack())
		}
	}()

	if !job.pool.Available() {
		time.Sleep(500 * time.Millisecond)
		return
	}

	rawMsg, ack, err := job.subject.FetchOne(ctx)
	if err != nil {
		job.errorf("tasks consumer fetch message error, %v", err)
		return
	}
	if rawMsg == nil {
		return
	}

	msg := rawMsg.(*subject.Task)
	taskId := msg.TaskId

	tasks, err := mongo.NewTask(job.Mongo).Find(ctx,
		&mongo.TaskFilter{
			TaskId: pointer.Of(taskId),
			Status: pointer.Of(models.TaskStatusQueue),
		},
	)
	if err != nil {
		job.errorf("query tasks %s error, %v", taskId, err)
		return
	}
	if len(tasks) == 0 {
		job.infof("task %s is not queuing, skip", taskId)
		if err = ack(); err != nil {
			job.errorf("acknowledges a message %s(skipped) error, %v", taskId, err)
		}
		return
	}

	if !job.pool.Add(1) {
		return
	}
	// 语言级别并发限制
	if constant.Types(msg.Types).HasSast() && !job.limiters.Add(msg.Lang) {
		job.pool.Done()
		return
	}

	if err = ack(); err != nil {
		job.errorf("acknowledges a message %s error, %v", taskId, err)
		job.pool.Done()
		job.limiters.Done(msg.Lang)
		return
	}
	if err = job.preHandle(ctx, tasks); err != nil {
		job.errorf("job previous handle %s error, %w", taskId, err)
		job.pool.Done()
		job.limiters.Done(msg.Lang)
		return
	}

	sig := NewLimiter(len(tasks))
	sig.Add(len(tasks))
	for i := range tasks {
		task := tasks[i]

		job.wg.Add(1)
		go func() {
			defer job.wg.Done()
			defer func() {
				if e := recover(); e != nil {
					job.errorf("run %s task %s panic: %v\n%s", task.Detail.Type, taskId, e, debug.Stack())
				}
				// 同时扫描sca、sast时，等待都结束时再pool done。
				if sig.DoneAndIsEmpty() {
					job.pool.Done()
				}
				if constant.TaskType(task.Detail.Type).IsSast() {
					job.limiters.Done(task.Detail.SastParams.Lang)
				}
			}()

			job.handle(ctx, &task)
		}()
	}
}

func (job *Job) handle(ctx context.Context, task *models.Task) {
	job.infof("%4s begin scanning task %s", task.Detail.Type, task.TaskId)

	ctxWithCancel, waitCancelCause := NewWaitCancelCause(ctx)
	key := task.TaskId + task.Detail.Type
	job.ctxCache.Store(key, waitCancelCause)

	now := time.Now()
	defer func() {
		if e := recover(); e != nil {
			task.Status = models.TaskStatusFailed
			task.ErrMsg = string(debug.Stack())
			job.errorf("task %s handle panic: %v\n%s", task.Id, e, task.ErrMsg)
		}

		if !utils.Canceled(ctxWithCancel) {
			// Note, context.Background() replaced ctx(can be canceled)
			if err := mongo.NewTask(job.Mongo).UpdateTask(context.Background(), task); err != nil {
				job.errorf("update task %s status error, %v", task.Id, err)
			}

			job.ctxCache.Delete(key)
		}

		waitCancelCause.Done()

		job.infof("%4s finish scanning task %s, cost %s", task.Detail.Type, task.TaskId, time.Since(now).String())
	}()

	// 修改从queueing 修改为 processing 任务状态,修改成功则继续
	task.Status = models.TaskStatusProcess
	modifyCount, err := mongo.NewTask(job.Mongo).UpdateTaskByFilter(ctx,
		bson.M{"task_id": task.TaskId, "detail.type": task.Detail.Type, "status": models.TaskStatusQueue},
		task,
	)
	if err != nil {
		job.errorf("update task %s error, %v", task.TaskId, err)
		return
	}
	if modifyCount == 0 {
		job.errorf("not updated task %s, and skip.", task.TaskId)
		return
	}

	var handler Handler
	handler, ok := job.handler[task.Detail.Type]
	if !ok {
		job.Logger.Errorf("get %s handler failed", task.Detail.Type)
		return
	}

	if err = handler.startJob(ctxWithCancel, task); err != nil {
		// 业务异常如解压所包失败等，请自行在starJob方法中修改task.ErrMsg
		task.Status = models.TaskStatusFailed
		task.ErrCode = dto.StatusInternalError
		task.DebugMsg = err.Error()
		return
	}

	if err = handler.processResult(ctxWithCancel, task); err != nil {
		job.Logger.Errorf("Process result failed: %v", err)
		task.Status = models.TaskStatusFailed
		task.ErrCode = dto.StatusInternalError
		task.ErrMsg = "服务内部故障"
		task.DebugMsg = err.Error()
	}
}

func (job *Job) runConfigConsumer(ctx context.Context) {
	job.wg.Add(1)
	go func(ctx context.Context) {
		defer job.wg.Done()

		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				job.debugf("config consumer context canceled")
				return
			case <-ticker.C:
				job.debugf("config consumer has been running")
			default:
			}
			job.configConsumer(ctx)
		}
	}(ctx)
}

func (job *Job) configConsumer(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			job.errorf("config consumer panic: %v\n%s", e, debug.Stack())
		}
	}()

	rawMsg, ack, err := job.configSubject.FetchOne(ctx)
	if err != nil {
		job.errorf("config consumer fetch message error, %v", err)
		return
	}
	if rawMsg == nil {
		return
	}
	if err = ack(); err != nil {
		job.errorf("acknowledges a config message error, %v", err)
		return
	}

	msg := rawMsg.(*subject.Config)
	job.pool.Tune(msg.Task.Concurrent)

	job.Config.Task.Concurrent = msg.Task.Concurrent
	job.Config.Task.ScaTimeout = utils.Sec2Duration(*msg.Task.ScaTimeout)
	job.Config.Task.SastTimeout = utils.Sec2Duration(*msg.Task.SastTimeout)
}

func (job *Job) runTerminating() {
	err := job.terminatingSubject.Subscribe(
		func(task *subject.Task) bool {
			defer func() {
				if e := recover(); e != nil {
					job.Logger.Error("terminating task subscribe panic: %v\n%s", e, debug.Stack())
				}
			}()

			var isTerminated bool
			for _, v := range constant.TaskTypes() {
				key := task.TaskId + v
				if value, ok := job.ctxCache.Load(key); ok {
					value.(*WaitCancelCause).WaitCanceled(errors.New("terminate"))
					job.ctxCache.Delete(key)
					isTerminated = true
					job.Logger.Debugf("terminated task %s", key)
				}
			}

			// 可达性分析
			if value, ok := job.ctxCache.Load(task.TaskId); ok {
				value.(*WaitCancelCause).WaitCanceled(errors.New("terminate"))
				job.ctxCache.Delete(task.TaskId)
				isTerminated = true
			}
			return isTerminated
		},
	)
	if err != nil {
		job.errorf("terminating consumer subscribe message error, %v", err)
		panic(err)
	}
}

func (job *Job) terminateProcessTasks() {
	// stop all processing task
	tasks, err := mongo.NewTask(job.Mongo).GetProcessingTasks(context.Background())
	if err != nil {
		job.Logger.Errorf("Get process task failed: %v", err)
		return
	}

	// clean all user detail info
	for _, task := range tasks {
		task.Status = models.TaskStatusFailed
		task.ErrCode = dto.StatusInternalError
		task.ErrMsg = "Stopped because server interrupt."
		if err = mongo.NewTask(job.Mongo).UpdateTask(context.TODO(), &task); err != nil {
			job.Logger.Errorf("Update task status failed: %v", err)
			return
		}
	}
}

func (job *Job) processTimeoutTasks(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			job.Logger.Error("process tasks timeout panic: %v\n%s", e, debug.Stack())
		}
	}()
	var tasks []models.Task
	{
		list, err := mongo.NewTask(job.Mongo).GetTimeoutTasks(ctx, constant.TypeSca, job.Config.Task.GetScaTimeout())
		if err != nil {
			job.Logger.Errorf("get timeout tasks error, %v", err)
			return
		}
		tasks = append(tasks, list...)
	}
	{
		list, err := mongo.NewTask(job.Mongo).GetTimeoutTasks(ctx, constant.TypeSast, job.Config.Task.GetSastTimeout())
		if err != nil {
			job.Logger.Errorf("get timeout tasks error, %v", err)
			return
		}
		tasks = append(tasks, list...)
	}
	{
		list, err := mongo.NewTask(job.Mongo).GetTimeoutTasks(ctx, constant.TypeBha, job.Config.Task.GetBhaTimeout())
		if err != nil {
			job.Logger.Errorf("get timeout tasks error, %v", err)
			return
		}
		tasks = append(tasks, list...)
	}

	if len(tasks) <= 0 {
		return
	}
	for _, task := range tasks {
		task.Status = models.TaskStatusFailed
		task.ErrCode = dto.StatusInternalError
		task.ErrMsg = "scanning timeout"
		if err := mongo.NewTask(job.Mongo).UpdateTask(ctx, &task); err != nil {
			job.Logger.Errorf("Update task %d status failed, %v", task.TaskId, err)
			return
		}
	}
}

func (job *Job) processDeletedModel(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			job.Logger.Error("Process delete model panic: %v\n%s", e, debug.Stack())
		}
	}()

	list, err := mongo.NewBhaModel(job.Mongo).GetDeleted(ctx)
	if err != nil {
		job.Logger.Errorf("Get deleted models error, %s", err)
		return
	}

	var tasks []models.Task
	for _, elem := range list {
		tasks, err = mongo.NewTask(job.Mongo).FindByFilter(ctx, bson.M{
			"status":          bson.M{"$in": models.TaskStatusQueuingAndProcessing()},
			"detail.type":     constant.TypeBha,
			"detail.model_id": elem.Id.Hex(),
		})
		if err != nil {
			job.Logger.Errorf("Find deleted models error, %s", err)
			return
		}

		// 没有任务，删除模型
		if len(tasks) == 0 {
			err = minio.New(job.Minio).RemoveObject(ctx, elem.Path)
			if err != nil {
				job.Logger.Errorf("Remove minio (%s)models error, %s", elem.Path, err)
				return
			}
			err = mongo.NewBhaModel(job.Mongo).Delete(ctx, elem.Id.Hex())
			if err != nil {
				job.Logger.Errorf("Remove (%s)models error, %s", elem.Path, err)
				return
			}
		}
	}
}

// 任务预处理
func (job *Job) preHandle(ctx context.Context, tasks []models.Task) (err error) {
	return nil
}

func (job *Job) debugf(template string, args ...interface{}) {
	job.Logger.Debugf(fmt.Sprintf("%s: %s", job.name, template), args...)
}

func (job *Job) infof(template string, args ...interface{}) {
	job.Logger.Infof(fmt.Sprintf("%s: %s", job.name, template), args...)
}

func (job *Job) errorf(template string, args ...interface{}) {
	job.Logger.Errorf(fmt.Sprintf("%s: %s", job.name, template), args...)
}
