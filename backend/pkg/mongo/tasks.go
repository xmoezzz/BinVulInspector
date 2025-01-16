package mongo

import (
	"context"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/pointer"
)

type Tasks struct {
	*base
}

func NewTasks(client *Client) *Tasks {
	return &Tasks{
		base: newBase(client, tasksCollection),
	}
}

type TasksFilter struct {
	name       *string
	source     *[]string
	taskId     *[]string
	skipTaskId *[]string

	status       *[]string
	detectMethod *string
	typ          *[]string
	startAt      *time.Time
	endAt        *time.Time

	sort struct {
		createdAt  *string
		modifiedAt *string
	}

	skip  *int64
	limit *int64
}

func (f *TasksFilter) SetName(name string) {
	f.name = &name
}

func (f *TasksFilter) SetSource(source ...string) {
	f.source = pointer.Of(source)
}

func (f *TasksFilter) SetTaskId(taskId ...string) {
	f.taskId = pointer.Of(taskId)
}

func (f *TasksFilter) SetSkipTaskId(skipTaskId ...string) {
	f.skipTaskId = pointer.Of(skipTaskId)
}

func (f *TasksFilter) SetStatus(status ...string) {
	f.status = pointer.Of(status)
}

func (f *TasksFilter) SetDetectMethod(method string) {
	f.detectMethod = pointer.Of(method)
}

func (f *TasksFilter) SetType(typ ...string) {
	f.typ = pointer.Of(typ)
}

func (f *TasksFilter) SetStartAt(startAt time.Time) {
	f.startAt = &startAt
}

func (f *TasksFilter) SetEndAt(endAt time.Time) {
	f.endAt = &endAt
}

func (f *TasksFilter) SetSortCreatedAt(sort string) {
	f.sort.createdAt = pointer.Of(sort)
}

func (f *TasksFilter) SetSkip(skip int64) {
	f.skip = &skip
}

func (f *TasksFilter) SetLimit(limit int64) {
	f.limit = &limit
}

// 聚合任务查询条件
// 默认相同task_id任务基础属性相同，故基础属性过滤放在前置过滤中
func (f *TasksFilter) query() (bson.M, bson.M) {
	previous, post := make(bson.M), make(bson.M)

	// 项目名称模糊查询
	if f.name != nil {
		if *f.name == "" {
			previous["name"] = ""
		} else {
			previous["name"] = bson.M{"$regex": regexp.QuoteMeta(*f.name)}
		}
	}

	// 任务来源
	if f.source != nil {
		previous["source"] = bson.M{
			"$in": in(*f.source),
		}
	}

	// 任务id查询
	if f.taskId != nil {
		previous["task_id"] = bson.M{
			"$in": in(*f.taskId),
		}
	}

	// 检测方式查询
	if f.detectMethod != nil {
		previous["detail.detection_method"] = *f.detectMethod
	}

	// 排除任务id查询
	if f.skipTaskId != nil {
		previous["task_id"] = bson.M{
			"$nin": in(*f.skipTaskId),
		}
	}

	// 状态过滤
	if f.status != nil {
		post["status"] = bson.M{
			"$in": in(*f.status),
		}
	}

	// 类型过滤
	if f.typ != nil {
		post["types"] = bson.M{
			"$in": in(*f.typ),
		}
	}

	// 时间过滤
	if f.startAt != nil && !f.startAt.IsZero() {
		post["created_at"] = bson.M{"$gte": *f.startAt}
	}
	if f.endAt != nil && !f.endAt.IsZero() {
		post["modified_at"] = bson.M{"$lte": *f.endAt}
	}

	return previous, post
}

func (f *TasksFilter) pipeline() mongo.Pipeline {
	previous, post := f.query()
	previousMatch := bson.D{
		{
			Key:   "$match",
			Value: previous,
		},
	}
	postMatch := bson.D{
		{
			Key:   "$match",
			Value: post,
		},
	}

	group := bson.D{
		{
			Key:   "$group",
			Value: groupFields(),
		},
	}
	project := bson.D{
		{
			Key:   "$project",
			Value: projectFields(),
		},
	}

	return mongo.Pipeline{previousMatch, group, project, postMatch}
}

func (f *TasksFilter) findPipeline(pipeline mongo.Pipeline) mongo.Pipeline {
	if len(pipeline) == 0 {
		pipeline = f.pipeline()
	}

	if f.sort.createdAt != nil || f.sort.modifiedAt != nil {
		sort := make(bson.M)
		if f.sort.createdAt != nil {
			sort["created_at"] = models.SortTypeValue(*f.sort.createdAt)
		}
		if f.sort.modifiedAt != nil {
			sort["modified_at"] = models.SortTypeValue(*f.sort.modifiedAt)
		}
		pipeline = append(pipeline, bson.D{
			{
				Key:   "$sort",
				Value: sort,
			},
		})
	}
	if f.skip != nil {
		pipeline = append(pipeline, bson.D{
			{
				Key:   "$skip",
				Value: *f.skip,
			},
		})
	}
	if f.limit != nil {
		pipeline = append(pipeline, bson.D{
			{
				Key:   "$limit",
				Value: *f.limit,
			},
		})
	}

	return pipeline
}

func groupFields() bson.M {
	return bson.M{
		"_id":           "$task_id",
		"name":          bson.M{"$first": "$name"},
		"source":        bson.M{"$first": "$source"},
		"description":   bson.M{"$first": "$description"},
		"file_path":     bson.M{"$first": "$file_path"},
		"file_hash":     bson.M{"$first": "$file_hash"},
		"file_size":     bson.M{"$max": "$file_size"},
		"types":         bson.M{"$addToSet": "$detail.type"},
		"status":        bson.M{"$addToSet": "$status"},
		"err_message":   bson.M{"$addToSet": "$err_message"},
		"created_at":    bson.M{"$min": "$created_at"},
		"modified_at":   bson.M{"$max": "$modified_at"},
		"debug_message": bson.M{"$addToSet": "$debug_message"},
		"detail":        bson.M{"$mergeObjects": "$detail"},
	}
}

func projectFields() bson.M {
	return bson.M{
		"_id":           1,
		"name":          1,
		"source":        1,
		"description":   1,
		"file_path":     1,
		"file_hash":     1,
		"file_size":     1,
		"types":         1,
		"err_message":   1,
		"created_at":    1,
		"modified_at":   1,
		"debug_message": 1,
		"status": bson.M{
			"$switch": bson.M{
				"branches": []bson.M{
					{
						"case": bson.M{
							"$in": []string{models.TaskStatusTerminated, "$status"},
						},
						"then": models.TaskStatusTerminated,
					},
					{
						"case": bson.M{
							"$in": []string{models.TaskStatusProcess, "$status"},
						},
						"then": models.TaskStatusProcess,
					},
					{
						"case": bson.M{
							"$in": []string{models.TaskStatusQueue, "$status"},
						},
						"then": models.TaskStatusQueue,
					},
					{
						"case": bson.M{
							"$in": []string{models.TaskStatusFailed, "$status"},
						},
						"then": models.TaskStatusFailed,
					},
				},
				"default": models.TaskStatusFinished,
			},
		},
		"detail": 1,
	}
}

func (c *Tasks) Find(ctx context.Context, filter *TasksFilter) ([]dto.TaskListItem, error) {
	pipeline := filter.findPipeline(nil)

	list, err := aggregate[dto.TaskListItem](ctx, c.collection(), pipeline)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (c *Tasks) List(ctx context.Context, filter *TasksFilter) (total int64, list []dto.TaskListItem, err error) {
	pipeline := filter.pipeline()

	// 查询数据总数
	{
		pipeline = append(pipeline, bson.D{
			{Key: "$count", Value: "total"},
		})

		type item struct {
			Total int64 `bson:"total"`
		}
		var items []item
		items, err = aggregate[item](ctx, c.collection(), pipeline)
		if err != nil {
			return 0, nil, err
		}
		if len(items) > 0 {
			total = items[0].Total
		}
		if total < 1 {
			return 0, nil, err
		}
	}

	// 查询数据
	{
		pipeline = pipeline[:len(pipeline)-1] // 去掉$count
		findPipeline := filter.findPipeline(pipeline)
		list, err = aggregate[dto.TaskListItem](ctx, c.collection(), findPipeline)
		if err != nil {
			return 0, nil, err
		}
	}

	return total, list, nil
}
