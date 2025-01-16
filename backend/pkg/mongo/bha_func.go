package mongo

import (
	"context"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/pointer"
)

type BhaFunc struct {
	*base
}

func NewBhaFunc(client *Client) *BhaFunc {
	return &BhaFunc{
		base: newBase(client, bhaFuncsCollection),
	}
}

func (c *BhaFunc) ListFunc(ctx context.Context, params dto.BhaFuncListReq) (total int64, list []models.BhaFunc, err error) {
	var filter bson.M
	{
		filter = bson.M{"task_id": params.TaskId}
		if params.Q != "" {
			filter["fname"] = bson.M{"$regex": regexp.QuoteMeta(params.Q), "$options": "i"}
		}
	}

	total, err = c.CountDocuments(ctx, filter)
	if err != nil {
		return 0, nil, err
	}

	findOptions := &options.FindOptions{
		Skip:  pointer.Of(params.Skip()),
		Limit: pointer.Of(params.PageSize),
	}

	if list, err = find[models.BhaFunc](ctx, c.collection(), filter, findOptions); err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

func (c *BhaFunc) DeleteByTaskIds(ctx context.Context, ids []string) (err error) {
	filter := bson.M{"task_id": bson.M{"$in": ids}}
	_, err = c.collection().DeleteMany(ctx, filter)
	return err
}
