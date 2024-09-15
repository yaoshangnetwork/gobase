package mqueue

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QueueItem struct {
	Channel string `bson:"channel"`
	Payload any    `bson:"payload"`
	Ack     string `bson:"ack"`
	Tries   int    `bson:"tries"`
}

func id() string {
	return uuid.New().String()
}

type MQueue struct {
	coll *mongo.Collection
	opts *QueueOpts
}

type Mode string

const (
	DebugMode   Mode = "debug"
	ReleaseMode Mode = "release"
)

type QueueOpts struct {
	Visibility time.Duration
	MaxTries   int
	Mode       Mode
}

func NewMQueue(db DBConfig, opts ...QueueOpts) *MQueue {
	opt := &QueueOpts{
		Visibility: time.Second * 60 * 5,
		MaxTries:   -1,
		Mode:       ReleaseMode,
	}

	if len(opts) > 0 {
		opt = &opts[0]
		if opt.Visibility <= 0 {
			panic("the value of visibility must be greater than 0")
		}
		if opt.Mode == "" {
			opt.Mode = ReleaseMode
		}
	}

	mq := &MQueue{coll: connect(db), opts: opt}
	mq.createIndexes()
	return mq
}

// createIndexes
func (mq *MQueue) createIndexes() {
	mq.coll.Indexes().DropAll(context.Background())

	_, err := mq.coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "visible", Value: 1},
			{Key: "dead", Value: 1},
			{Key: "deleted", Value: 1},
		},
	})
	if err != nil {
		panic(err)
	}

	_, err = mq.coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "ack", Value: 1},
			{Key: "visible", Value: 1},
			{Key: "dead", Value: 1},
			{Key: "deleted", Value: 1},
		},
	})
	if err != nil {
		panic(err)
	}

	// ack 唯一索引
	_, err = mq.coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "ack", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		panic(err)
	}

	// 已完成的 queue item 保留 7 天数据
	_, err = mq.coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "deleted", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(3600 * 24 * 7),
	})
	if err != nil {
		panic(err)
	}
}

type Item struct {
	Channel string
	Payload any
	Delay   time.Duration
}

// Clean
func (mq *MQueue) Clean() {
	if mq.opts.Mode != DebugMode {
		log.Print("the \"Clean\" method can only be used in debug mode")
		return
	}
	mq.coll.DeleteMany(context.Background(), bson.M{})
}

// Add
func (mq *MQueue) Add(item ...Item) bool {
	return mq.insertItems(item)
}

func (mq *MQueue) insertItems(items []Item) bool {
	if len(items) == 0 {
		return false
	}
	docs := make([]interface{}, 0, len(items))
	for _, item := range items {
		doc := map[string]any{
			"ack":     id(),
			"payload": item.Payload,
			"visible": time.Now().Add(item.Delay),
			"tries":   int(0),
			"dead":    false,
		}
		if item.Channel != "" {
			doc["channel"] = item.Channel
		}
		docs = append(docs, doc)
	}
	_, err := mq.coll.InsertMany(context.Background(), docs)
	return err == nil
}

// Get
func (mq *MQueue) Get() (*QueueItem, bool) {
	query := bson.M{"visible": bson.M{"$lte": time.Now()}, "dead": false, "deleted": nil}
	update := bson.M{
		"$inc": bson.M{"tries": 1},
		"$set": bson.M{"ack": id(), "visible": time.Now().Add(mq.opts.Visibility)},
	}
	after := options.After
	res := mq.coll.FindOneAndUpdate(context.Background(), query, update, &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	})
	err := res.Err()
	if err != nil {
		return nil, false
	}
	item := new(QueueItem)
	err = res.Decode(item)
	if err != nil {
		return nil, false
	}
	if mq.opts.MaxTries > 0 && item.Tries > mq.opts.MaxTries {
		mq.coll.UpdateOne(
			context.Background(),
			bson.M{"ack": item.Ack},
			bson.M{"$set": bson.M{"dead": true}},
		)
		return nil, false
	}
	return item, true
}

// Ack
func (mq *MQueue) Ack(ack string) bool {
	query := bson.M{"ack": ack, "visible": bson.M{"$gt": time.Now()}, "dead": false, "deleted": nil}
	update := bson.M{
		"$set": bson.M{"deleted": time.Now()},
	}
	res, _ := mq.coll.UpdateOne(context.Background(), query, update)
	return res.ModifiedCount == 1
}

// Ping
func (mq *MQueue) Ping(ack string) bool {
	query := bson.M{"ack": ack, "visible": bson.M{"$gt": time.Now()}, "dead": false, "deleted": nil}
	update := bson.M{
		"$set": bson.M{"visible": time.Now().Add(mq.opts.Visibility)},
	}
	res, err := mq.coll.UpdateOne(context.Background(), query, update)
	if err != nil {
		return false
	}
	return res.MatchedCount == 1
}

// Total
func (mq *MQueue) Total() (int64, error) {
	return mq.coll.CountDocuments(context.Background(), bson.M{})
}

// Size
func (mq *MQueue) Size() (int64, error) {
	query := bson.M{"visible": bson.M{"$lte": time.Now()}, "dead": false, "deleted": nil}
	return mq.coll.CountDocuments(context.Background(), query)
}

// InFlight
func (mq *MQueue) InFlight() (int64, error) {
	query := bson.M{"visible": bson.M{"$gt": time.Now()}, "dead": false, "deleted": nil}
	return mq.coll.CountDocuments(context.Background(), query)
}

// Done
func (mq *MQueue) Done() (int64, error) {
	query := bson.M{"deleted": bson.M{"$exists": true}}
	return mq.coll.CountDocuments(context.Background(), query)
}

// Dead
func (mq *MQueue) Dead() (int64, error) {
	query := bson.M{"dead": true}
	return mq.coll.CountDocuments(context.Background(), query)
}

// PayloadInt32
func (item *QueueItem) PayloadInt32() (int32, bool) {
	v, ok := item.Payload.(int32)
	return v, ok
}

// PayloadInt64
func (item *QueueItem) PayloadInt64() (int64, bool) {
	v, ok := item.Payload.(int64)
	return v, ok
}

// PayloadFloat64
func (item *QueueItem) PayloadFloat64() (float64, bool) {
	v, ok := item.Payload.(float64)
	return v, ok
}

// PayloadString
func (item *QueueItem) PayloadString() (string, bool) {
	v, ok := item.Payload.(string)
	return v, ok
}

// PayloadBool
func (item *QueueItem) PayloadBool() (bool, bool) {
	v, ok := item.Payload.(bool)
	return v, ok
}

// PayloadDecode
func (item *QueueItem) PayloadDecode(doc any) bool {
	switch item.Payload.(type) {
	case primitive.D:
		return decode(item.Payload, doc)
	case primitive.A:
		return decodeSlice(item.Payload, doc)
	}
	return false
}

func decode(data any, doc any) bool {
	m, err := bson.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if err = bson.Unmarshal(m, doc); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func decodeSlice(data any, doc any) bool {
	// 检查 doc 是否是指向切片的指针
	val := reflect.ValueOf(doc)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return false
	}
	for _, item := range data.(primitive.A) {
		b, err := bson.Marshal(item)
		if err != nil {
			return false
		}
		elem := reflect.New(val.Elem().Type().Elem()).Interface()
		if err := bson.Unmarshal(b, elem); err != nil {
			return false
		}
		val.Elem().Set(reflect.Append(val.Elem(), reflect.ValueOf(elem).Elem()))
	}
	return true
}
