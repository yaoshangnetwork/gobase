package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IBaseRepo[T any] interface {
	InsertOne(doc *T) (primitive.ObjectID, error)
	InsertMany(docs []*T) ([]primitive.ObjectID, error)

	FindOne(filter bson.M) (*T, error)
	FindByID(id primitive.ObjectID) (*T, error)

	UpdateOne(filter bson.M, update any) error
	UpdateByID(id primitive.ObjectID, update any) error

	List(filter bson.M, page int64, size int64) ([]*T, int64, error)
	All(filter bson.M) ([]*T, error)

	Count(filter bson.M) (int64, error)
	DeleteByID(id primitive.ObjectID) error
}

type BaseRepo[T any] struct {
	Coll *mongodb.Collection
}

var _ IBaseRepo[any] = (*BaseRepo[any])(nil)

// nowTimestamp 当时时间的毫秒级时间戳
func nowTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (r *BaseRepo[T]) InsertOne(doc *T) (primitive.ObjectID, error) {
	result, err := r.Coll.InsertOne(context.Background(), doc)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id := result.InsertedID.(primitive.ObjectID)
	return id, nil
}

func (r *BaseRepo[T]) InsertMany(docs []*T) ([]primitive.ObjectID, error) {
	// 转换为 []interface{}
	interfaceDocs := make([]interface{}, len(docs))
	for i, v := range docs {
		interfaceDocs[i] = v
	}

	result, err := r.Coll.InsertMany(context.Background(), interfaceDocs)
	if err != nil {
		return nil, err
	}

	resultIDs := make([]primitive.ObjectID, len(result.InsertedIDs))
	for i, v := range result.InsertedIDs {
		resultIDs[i] = v.(primitive.ObjectID)
	}

	return resultIDs, nil
}

func (r *BaseRepo[T]) FindOne(filter bson.M) (*T, error) {
	result := new(T)
	if filter == nil {
		filter = bson.M{}
	}
	filter["is_deleted"] = false
	err := r.Coll.FindOne(context.Background(), filter).Decode(result)
	return result, err
}

func (r *BaseRepo[T]) FindByID(id primitive.ObjectID) (*T, error) {
	filter := bson.M{"_id": id}
	return r.FindOne(filter)
}

func (r *BaseRepo[T]) UpdateOne(filter bson.M, update any) error {
	if filter == nil {
		filter = bson.M{}
	}
	filter["is_deleted"] = false
	_, err := r.Coll.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	return err
}

func (r *BaseRepo[T]) UpdateByID(id primitive.ObjectID, update any) error {
	return r.UpdateOne(bson.M{"_id": id}, update)
}

func (r *BaseRepo[T]) List(filter bson.M, page int64, size int64) ([]*T, int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["is_deleted"] = false

	result := make([]*T, 0, size)

	cursor, err := r.Coll.Find(
		context.Background(),
		filter,
		options.Find().SetSkip((page-1)*size).SetLimit(size),
	)
	if err != nil {
		return result, 0, err
	}
	defer cursor.Close(context.Background())
	if err = cursor.All(context.Background(), result); err != nil {
		return result, 0, err
	}

	count, err := r.Count(filter)
	if err != nil {
		return result, 0, err
	}

	return result, count, nil
}

func (r *BaseRepo[T]) All(filter bson.M) ([]*T, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["is_deleted"] = false

	result := make([]*T, 0)

	cursor, err := r.Coll.Find(
		context.Background(),
		filter,
	)
	if err != nil {
		return result, err
	}
	defer cursor.Close(context.Background())
	if err = cursor.All(context.Background(), result); err != nil {
		return result, err
	}

	return result, nil
}

func (r *BaseRepo[T]) Count(filter bson.M) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["is_deleted"] = false
	return r.Coll.CountDocuments(context.Background(), filter)
}

func (r *BaseRepo[T]) DeleteByID(id primitive.ObjectID) error {
	_, err := r.Coll.UpdateOne(
		context.Background(),
		bson.M{"_id": id, "is_deleted": false},
		bson.M{
			"$set": bson.M{"deleted_at": nowTimestamp(), "is_deleted": true},
		})
	return err
}
