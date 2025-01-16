package mongo

import (
	"context"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/pointer"
)

type BhaModel struct {
	*base
}

func NewBhaModel(client *Client) *BhaModel {
	return &BhaModel{
		base: newBase(client, bhaModelsCollection),
	}
}

func (c *BhaModel) FindOneByNameAndType(ctx context.Context, name string, typ string) (m *models.BhaModel, err error) {
	filter := bson.M{"name": name, "type": typ}

	return findOne[models.BhaModel](ctx, c.collection(), filter)
}

func (c *BhaModel) FindById(ctx context.Context, id string) (m *models.BhaModel, err error) {
	return findById[models.BhaModel](ctx, c.collection(), id)
}

func (c *BhaModel) UpsertBhaModelByPath(ctx context.Context, path string, m *models.BhaModel) (err error) {
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: "path", Value: path}}
	update := bson.D{
		{Key: "$set", Value: m},
		{Key: "$setOnInsert", Value: bson.M{"created_at": time.Now()}},
	}

	_, err = c.collection().UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func (c *BhaModel) List(ctx context.Context, params dto.BhaModelListReq) (total int64, list []models.BhaModel, err error) {
	filter := bson.M{}
	{
		if params.Name != "" {
			filter["name"] = bson.M{"$regex": regexp.QuoteMeta(params.Name), "$options": "i"}
		}
		if len(params.Types) > 0 {
			filter["type"] = bson.M{"$in": params.Types}
		}
		if !params.ShowDeleted {
			filter["deleted_at"] = bson.M{"$exists": false}
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

	if list, err = find[models.BhaModel](ctx, c.collection(), filter, findOptions); err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

func (c *BhaModel) GetDeleted(ctx context.Context) (list []models.BhaModel, err error) {
	filter := bson.M{"deleted_at": bson.M{"$exists": true}}
	return find[models.BhaModel](ctx, c.collection(), filter)
}
