package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/models"
)

type Task struct {
	*base
}

func NewTask(client *Client) *Task {
	return &Task{
		base: newBase(client, tasksCollection),
	}
}

func (c *Task) ExpiredList(ctx context.Context, expiredAt time.Time) (list []models.Task, err error) {
	// todo 需要聚合后再查询
	filter := bson.M{
		"status": bson.M{
			"$in": []string{models.TaskStatusFinished, models.TaskStatusFailed, models.TaskStatusTerminated},
		},
		"created_at": bson.M{
			"$lt": expiredAt,
		},
	}
	findOptions := &options.FindOptions{
		Sort: bson.D{
			{Key: "created_at", Value: models.Asc},
		},
	}
	findOptions.SetBatchSize(BatchSize)

	{
		var cursor *mongo.Cursor
		if cursor, err = c.collection().Find(ctx, filter, findOptions); err != nil {
			return nil, err
		}
		defer func() { _ = cursor.Close(ctx) }()

		if err = cursor.All(ctx, &list); err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (c *Task) DeleteByTaskIds(ctx context.Context, ids []string) (err error) {
	filter := bson.M{
		"task_id": bson.M{
			"$in": ids,
		},
	}
	_, err = c.collection().DeleteMany(ctx, filter)
	return err
}

func (c *Task) GetProcessingTasks(ctx context.Context) ([]models.Task, error) {
	filter := bson.M{"status": models.TaskStatusProcess}
	return find[models.Task](ctx, c.collection(), filter)
}

func (c *Task) GetTimeoutTasks(ctx context.Context, taskType string, duration time.Duration) ([]models.Task, error) {
	filter := bson.M{
		"status":      models.TaskStatusQueue,
		"detail.type": taskType,
		"detail.stage": bson.M{
			"$eq": models.TaskStagePending,
		},
		"created_at": bson.M{
			"$lt": time.Now().Add(-1 * duration),
		},
	}
	return find[models.Task](ctx, c.collection(), filter)
}

func (c *Task) findOne(ctx context.Context, filter bson.M, opts ...*options.FindOneOptions) (*models.Task, error) {
	return findOne[models.Task](ctx, c.collection(), filter, opts...)
}

func (c *Task) GetSastTask(ctx context.Context, taskId string) (*models.Task, error) {
	return c.findOne(ctx, bson.M{"task_id": taskId, "detail.type": constant.TypeSast})
}

func (c *Task) GetScaTask(ctx context.Context, taskId string) (*models.Task, error) {
	return c.findOne(ctx, bson.M{"task_id": taskId, "detail.type": constant.TypeSca})
}

func (c *Task) GetBhaTask(ctx context.Context, taskId string) (*models.Task, error) {
	return c.findOne(ctx, bson.M{"task_id": taskId, "detail.type": constant.TypeBha})
}

func (c *Task) GetTasksByTaskId(ctx context.Context, taskId string) ([]models.Task, error) {
	return find[models.Task](ctx, c.collection(), bson.M{"task_id": taskId})
}

func (c *Task) FindByUserIdAndTaskId(ctx context.Context, userId, taskId string) (*models.Task, error) {
	filter := bson.M{
		"task_id": taskId,
		"user_id": userId,
	}
	return c.findOne(ctx, filter)
}

func (c *Task) FindByTaskIdAndType(ctx context.Context, taskId, taskType string) (*models.Task, error) {
	filter := bson.M{
		"task_id":     taskId,
		"detail.type": taskType,
	}
	return c.findOne(ctx, filter)
}

func (c *Task) UpdateTask(ctx context.Context, task *models.Task) error {
	filter := bson.M{"task_id": task.TaskId, "detail.type": task.Detail.Type}
	_, err := c.UpdateTaskByFilter(ctx, filter, task)
	return err
}

func (c *Task) UpdateTaskByFilter(ctx context.Context, filter interface{}, task *models.Task) (count int64, err error) {
	update := bson.M{
		"$set": bson.M{
			"detail":        task.Detail,
			"status":        task.Status,
			"result":        task.Result,
			"debug_message": task.DebugMsg,
			"err_code":      task.ErrCode,
			"err_message":   task.ErrMsg,
			"modified_at":   time.Now(),

			// 扫描镜像时, 添加镜像地址
			"file_hash": task.FileHash,
			"file_path": task.FilePath,
			"file_size": task.FileSize,
		},
	}
	updateResult, err := c.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return updateResult.ModifiedCount, nil
}

func (c *Task) InsertMany(ctx context.Context, documents []models.Task) (err error) {
	_, err = insertMany(ctx, c.collection(), documents)
	return err
}

func (c *Task) ResetStatus(ctx context.Context, taskId string) error {
	filter := bson.M{
		"task_id": taskId,
	}
	fields := bson.M{
		"$set": bson.M{
			"status":       models.TaskStatusQueue,
			"detail.stage": "",
		},
	}
	_, err := c.collection().UpdateMany(ctx, filter, fields)

	return err
}

func (c *Task) UpdateUploadFileByTaskId(ctx context.Context, taskId string, file dto.UploadFile) (err error) {
	filter := bson.M{"task_id": taskId}
	fields := bson.M{
		"$set": bson.M{
			"file_path": file.FilePath,
			"file_hash": file.FileHash,
			"file_size": file.FileSize,
		},
	}
	_, err = c.collection().UpdateMany(ctx, filter, fields)
	return err
}

type TaskFilter struct {
	TaskId *string `bson:"task_id,omitempty"`
	Status *string `bson:"status,omitempty"`
}

func (c *Task) Find(ctx context.Context, filter *TaskFilter) ([]models.Task, error) {
	return find[models.Task](ctx, c.collection(), filter)
}

func (c *Task) FindByFilter(ctx context.Context, filter interface{}) ([]models.Task, error) {
	return find[models.Task](ctx, c.collection(), filter)
}

type TaskFields struct {
	Status *string `bson:"status,omitempty"`
}

func (c *Task) UpdateMany(ctx context.Context, filter *TaskFilter, fields *TaskFields) (*mongo.UpdateResult, error) {
	update := bson.M{
		"$set": fields,
	}
	return c.collection().UpdateMany(ctx, filter, update)
}
