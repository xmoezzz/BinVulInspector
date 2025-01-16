package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	BatchSize = 50
)

type Filter interface {
	Filter() bson.M
}

type base struct {
	client *mongo.Client
	db     string
	coll   string
}

type Option func(*base)

func withDB(db string) Option {
	return func(b *base) {
		b.db = db
	}
}

func newBase(client *Client, collection string, opts ...Option) *base {
	o := &base{
		db:     client.AuthSource,
		client: client.client,
		coll:   collection,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *base) database() *mongo.Database {
	return o.client.Database(o.db)
}

func (o *base) collection() *mongo.Collection {
	return o.database().Collection(o.coll)
}

func (o *base) Insert(ctx context.Context, document interface{}) (string, error) {
	result, err := o.collection().InsertOne(ctx, document)
	if err != nil {
		return "", err
	}
	return o.Id(result.InsertedID)
}

func (o *base) Id(value interface{}) (string, error) {
	oid, ok := value.(primitive.ObjectID)
	if !ok {
		return "", errors.New("value not type primitive.ObjectID")
	}
	return oid.Hex(), nil
}

func (o *base) InsertMany(ctx context.Context, documents []interface{}) (err error) {
	_, err = o.collection().InsertMany(ctx, documents)
	return err
}

func (o *base) Update(ctx context.Context, id string, document interface{}) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = o.collection().UpdateByID(ctx, objID, bson.D{{Key: "$set", Value: document}})
	return err
}

func (o *base) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{
		"_id": objID,
	}
	_, err = o.collection().DeleteOne(ctx, filter)
	return err
}

func (o *base) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	return o.collection().CountDocuments(ctx, filter)
}

func (o *base) CountDocumentsWithPipeline(ctx context.Context, pipeline mongo.Pipeline) (total int64, err error) {
	type item struct {
		Total int64 `bson:"total"`
	}

	var items []item
	items, err = aggregate[item](ctx, o.collection(), append(pipeline, bson.D{{Key: "$count", Value: "total"}}))
	if err != nil {
		return 0, err
	}
	if len(items) > 0 {
		total = items[0].Total
	}
	if total < 1 {
		total = 0
	}
	return total, nil
}

type UpdateResult struct {
	*mongo.UpdateResult
}

func findOne[T any](ctx context.Context, col *mongo.Collection, filter interface{}, opts ...*options.FindOneOptions) (m *T, err error) {
	if err = col.FindOne(ctx, filter, opts...).Decode(&m); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, nil
	}
	return m, nil
}

func findById[T any](ctx context.Context, col *mongo.Collection, id string) (m *T, err error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{
		"_id": objID,
	}
	return findOne[T](ctx, col, filter)
}

func findByUserIdAndId[T any](ctx context.Context, col *mongo.Collection, userId, id string) (m *T, err error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{
		"_id":     objID,
		"user_id": userId,
	}
	return findOne[T](ctx, col, filter)
}

func insertMany[T any](ctx context.Context, col *mongo.Collection, documents []T) ([]primitive.ObjectID, error) {
	docs := make([]interface{}, len(documents))
	for i := range documents {
		docs[i] = documents[i]
	}
	result, err := col.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, len(documents))
	for i := range result.InsertedIDs {
		ids[i] = result.InsertedIDs[i].(primitive.ObjectID)
	}
	return ids, nil
}

func find[T any](ctx context.Context, col *mongo.Collection, filter interface{}, opts ...*options.FindOptions) (list []T, err error) {
	opt := &options.FindOptions{}
	opt.SetBatchSize(BatchSize)
	opts = append(opts, opt)

	cur, err := col.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()

	if err = cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func aggregate[T any](ctx context.Context, col *mongo.Collection, pipeline mongo.Pipeline, opts ...*options.AggregateOptions) (list []T, err error) {
	opt := &options.AggregateOptions{}
	opt.SetBatchSize(BatchSize)
	opts = append(opts, opt)

	var cursor *mongo.Cursor
	if cursor, err = col.Aggregate(ctx, pipeline, opts...); err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func ObjectID(id string) primitive.ObjectID {
	oid, _ := primitive.ObjectIDFromHex(id)
	return oid
}

func ObjectIDWithError(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

func ObjectIDs(ids []string) []primitive.ObjectID {
	oids := make([]primitive.ObjectID, len(ids))
	for i := range ids {
		oid, _ := primitive.ObjectIDFromHex(ids[i])
		oids[i] = oid
	}
	return oids
}

func in[T any](slice []T) []T {
	// $in needs an array
	return append([]T{}, slice...)
}
