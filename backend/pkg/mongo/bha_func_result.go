package mongo

import (
	"context"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/pointer"
)

type BhaFuncResult struct {
	*base
}

func NewBhaFuncResult(client *Client) *BhaFuncResult {
	return &BhaFuncResult{
		base: newBase(client, bhaFuncResultsCollection),
	}
}

func (c *BhaFuncResult) ListFuncResult(ctx context.Context, params dto.BhaFuncResultListReq) (total int64, list []models.BhaFuncResult, err error) {
	var filter bson.M
	{
		filter = bson.M{"task_id": params.TaskId, "func_id": params.FuncId}
	}

	var pipeline mongo.Pipeline
	{
		match := bson.D{{Key: "$match", Value: filter}}
		sort := bson.D{{Key: "$sort", Value: bson.M{"sim": models.Desc}}}

		pipeline = mongo.Pipeline{match, sort}

		if !pointer.IsNil(params.TopN) {
			pipeline = append(pipeline, bson.D{{Key: "$limit", Value: pointer.PAny(params.TopN)}})
		}
		if params.Q != "" {
			postFilter := bson.M{"fname": bson.M{"$regex": regexp.QuoteMeta(params.Q), "$options": "i"}}
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: postFilter}})
		}
	}

	skip := bson.D{{Key: "$skip", Value: params.Skip()}}
	limit := bson.D{{Key: "$limit", Value: params.PageSize}}

	if total, err = c.CountDocumentsWithPipeline(ctx, pipeline); err != nil {
		return 0, nil, err
	}

	if list, err = aggregate[models.BhaFuncResult](ctx, c.collection(), append(pipeline, skip, limit)); err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

func (c *BhaFuncResult) DeleteByTaskIds(ctx context.Context, ids []string) (err error) {
	filter := bson.M{"task_id": bson.M{"$in": ids}}
	_, err = c.collection().DeleteMany(ctx, filter)
	return err
}
