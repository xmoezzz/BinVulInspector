package mongo

import (
	"bytes"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/utils"
)

func InitCollections(ctx context.Context, client *Client) (err error) {
	var collections []string
	collections, err = newBase(client, "").database().ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("get database collections error, %w", err)
	}

	collectionsIndices := collectionsAndIndices()
	for name, indices := range collectionsIndices {
		if !utils.Contains(collections, name) {
			if err = newBase(client, "").database().CreateCollection(ctx, name); err != nil {
				return fmt.Errorf("create  collection %s error, %w", name, err)
			}
		}
		if len(indices) <= 0 {
			continue
		}

		indexView := newBase(client, name).collection().Indexes()
		var items []*mongo.IndexSpecification

		if items, err = indexView.ListSpecifications(ctx); err != nil {
			return fmt.Errorf("collection %s listSpecifications error, %w", name, err)
		}
		for i := range indices {
			var keysDocument []byte
			keysDocument, err = bson.Marshal(indices[i].Keys)
			if err != nil {
				return err
			}
			f := func(item *mongo.IndexSpecification) bool {
				return bytes.Equal(item.KeysDocument, keysDocument)
			}
			if utils.ContainsFunc(items, f) {
				continue
			}

			// 处理text索引，只能存在一个， text索引一个集合中只能存在一个
			if isTextIndex(keysDocument) {
				// 删除旧的text索引
				for _, item := range items {
					if item.KeysDocument.Lookup("_fts").String() == `"text"` {
						if _, err = indexView.DropOne(ctx, item.Name); err != nil {
							return fmt.Errorf(
								"delete collection %s index %s error, %w",
								name, item.Name, err,
							)
						}
						break
					}
				}
			}

			if _, err = indexView.CreateOne(ctx, indices[i]); err != nil {
				return fmt.Errorf(
					"create collection %s index %s error, %w",
					name, *indices[i].Options.Name, err,
				)
			}
		}
	}

	return nil
}

func collectionsAndIndices() map[string][]mongo.IndexModel {
	return map[string][]mongo.IndexModel{
		tasksCollection: {
			{
				Keys: bson.D{{Key: "task_id", Value: models.Asc}},
			},
			{
				Keys: bson.D{{Key: "created_at", Value: models.Desc}},
			},
			{
				Keys: bson.D{{Key: "name", Value: models.Asc}},
			},
		},
		bhaFuncsCollection: {
			{
				Keys: bson.D{{Key: "task_id", Value: models.Asc}},
			},
			{
				Keys: bson.D{{Key: "file_id", Value: models.Asc}},
			},
		},
		bhaFuncResultsCollection: {
			{
				Keys: bson.D{{Key: "task_id", Value: models.Asc}},
			},
			{
				Keys: bson.D{{Key: "func_id", Value: models.Asc}},
			},
			{
				Keys: bson.D{
					{Key: "task_id", Value: models.Asc},
					{Key: "func_id", Value: models.Asc},
					{Key: "sim", Value: models.Desc},
				},
			},
		},
		bhaModelsCollection: {},
		configsCollection:   {},
	}
}

func isTextIndex(keysDocument []byte) bool {
	rawValues, err := bson.Raw(keysDocument).Values()
	if err != nil {
		return false
	}
	for _, rawValue := range rawValues {
		if rawValue.String() == `"text"` {
			return true
		}
	}
	return false
}
