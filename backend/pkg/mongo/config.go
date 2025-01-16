package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"bin-vul-inspector/pkg/models"
)

type Config struct {
	*base
}

func NewConfig(client *Client) *Config {
	return &Config{
		base: newBase(client, configsCollection),
	}
}

func (c *Config) Latest(ctx context.Context) (config *models.Config, err error) {
	findOptions := &options.FindOneOptions{
		Sort: bson.D{{Key: "_id", Value: models.Desc}},
	}
	if err = c.collection().FindOne(ctx, bson.M{}, findOptions).Decode(&config); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, nil
	}
	return config, nil
}
