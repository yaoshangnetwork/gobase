package mongo

import (
	"context"
	"errors"
	"reflect"
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

	DeleteOne(filter bson.M) error
	DeleteByID(id primitive.ObjectID) error

	ForceDeleteOne(filter bson.M) error
	ForceDeleteByID(id primitive.ObjectID) error

	List(filter bson.M, page int64, size int64) ([]*T, int64, error)
	All(filter bson.M) ([]*T, error)
	Exist(filter bson.M) (bool, error)
	Count(filter bson.M) (int64, error)

	WithContext(ctx context.Context) IBaseRepo[T]
}

type BaseRepo[T any] struct {
	ctx  *context.Context
	Coll *mongodb.Collection
}

var _ IBaseRepo[any] = (*BaseRepo[any])(nil)

// nowTimestamp 当时时间的毫秒级时间戳
func nowTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (r *BaseRepo[T]) getContext() context.Context {
	if r.ctx != nil {
		return *r.ctx
	}
	return context.Background()
}

func (r *BaseRepo[T]) InsertOne(doc *T) (primitive.ObjectID, error) {
	v := reflect.ValueOf(doc)

	method := v.MethodByName("BeforeCreate")
	if method.IsValid() && method.Type().NumIn() == 0 && method.Type().NumOut() == 0 {
		method.Call(nil)
	}

	result, err := r.Coll.InsertOne(r.getContext(), doc)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id := result.InsertedID.(primitive.ObjectID)
	return id, nil
}

func (r *BaseRepo[T]) InsertMany(docs []*T) ([]primitive.ObjectID, error) {
	// 转换为 []interface{}
	interfaceDocs := make([]interface{}, len(docs))
	for i, doc := range docs {
		v := reflect.ValueOf(doc)
		method := v.MethodByName("BeforeCreate")
		if method.IsValid() && method.Type().NumIn() == 0 && method.Type().NumOut() == 0 {
			method.Call(nil)
		}
		interfaceDocs[i] = doc
	}

	result, err := r.Coll.InsertMany(r.getContext(), interfaceDocs)
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
	err := r.Coll.FindOne(r.getContext(), filter).Decode(result)
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
		r.getContext(),
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
		r.getContext(),
		filter,
		options.Find().SetSkip((page-1)*size).SetLimit(size),
	)
	if err != nil {
		return result, 0, err
	}
	defer cursor.Close(r.getContext())
	if err = cursor.All(r.getContext(), result); err != nil {
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
		r.getContext(),
		filter,
	)
	if err != nil {
		return result, err
	}
	defer cursor.Close(r.getContext())
	if err = cursor.All(r.getContext(), result); err != nil {
		return result, err
	}

	return result, nil
}

func (d *BaseRepo[T]) Exist(filter bson.M) (bool, error) {
	_, err := d.FindOne(filter)
	if err != nil {
		if errors.Is(err, mongodb.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *BaseRepo[T]) Count(filter bson.M) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["is_deleted"] = false
	return r.Coll.CountDocuments(r.getContext(), filter)
}

func (r *BaseRepo[T]) DeleteOne(filter bson.M) error {
	if filter == nil {
		filter = bson.M{}
	}
	filter["is_deleted"] = false
	_, err := r.Coll.UpdateOne(
		r.getContext(),
		filter,
		bson.M{
			"$set": bson.M{"deleted_at": nowTimestamp(), "is_deleted": true},
		})
	return err
}

func (r *BaseRepo[T]) DeleteByID(id primitive.ObjectID) error {
	return r.DeleteOne(bson.M{"_id": id})
}

func (r *BaseRepo[T]) ForceDeleteOne(filter bson.M) error {
	_, err := r.Coll.DeleteOne(r.getContext(), filter)
	return err
}

func (r *BaseRepo[T]) ForceDeleteByID(id primitive.ObjectID) error {
	_, err := r.Coll.DeleteOne(r.getContext(), bson.M{"_id": id})
	return err
}

func (r *BaseRepo[T]) WithContext(ctx context.Context) IBaseRepo[T] {
	return &BaseRepo[T]{
		ctx:  &ctx,
		Coll: r.Coll,
	}
}
