package mqueue

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageNode struct {
	coll    *mongo.Collection
	mode    Mode
	channel *ChannelNode
}

type IMessageNode interface {
	Add(message ...Message) bool
	Clear(channel ...string) bool // dev mode only
	Get(channel string) (*QueueMessage, bool)
	Watch(channel string, interval time.Duration) chan *QueueMessage
	Ack(ack string) bool
	Ping(channel, ack string) bool
	Total(channel ...string) (int64, error)
	Size(channel ...string) (int64, error)
	InFlight(channel ...string) (int64, error)
	Done(channel ...string) (int64, error)
	Dead(channel ...string) (int64, error)
}

type QueueMessage struct {
	Channel string `bson:"channel"`
	Message any    `bson:"message"`
	Ack     string `bson:"ack"`
	Tries   int    `bson:"tries"`
}

type Message struct {
	Channel string
	Message any
	Delay   time.Duration
}

func id() string {
	return uuid.New().String()
}

var _ IMessageNode = (*MessageNode)(nil)

// Add
func (msg *MessageNode) Add(message ...Message) bool {
	if len(message) == 0 {
		return false
	}

	for _, m := range message {
		if m.Channel == "" {
			return false
		}
		if m.Message == nil {
			return false
		}
	}

	docs := make([]interface{}, 0, len(message))
	for _, item := range message {
		_, ok := msg.channel.Get(item.Channel)
		if !ok {
			return false
		}
		doc := map[string]any{
			"channel": item.Channel,
			"ack":     id(),
			"message": item.Message,
			"visible": time.Now().Add(item.Delay),
			"tries":   int(0),
			"dead":    false,
		}
		docs = append(docs, doc)
	}
	_, err := msg.coll.InsertMany(context.Background(), docs)
	return err == nil
}

// Clear
func (msg *MessageNode) Clear(channel ...string) bool {
	if msg.mode != DebugMode {
		log.Print("the \"Clean\" method can only be used in debug mode")
		return false
	}
	if len(channel) >= 1 {
		_, err := msg.coll.DeleteMany(context.Background(), bson.M{"channel": channel[0]})
		return err == nil
	}
	_, err := msg.coll.DeleteMany(context.Background(), bson.M{})
	return err == nil
}

// Get
func (msg *MessageNode) Get(channel string) (*QueueMessage, bool) {
	c, ok := msg.channel.Get(channel)
	if !ok {
		return nil, false
	}
	query := bson.M{"channel": channel, "visible": bson.M{"$lte": time.Now()}, "dead": false, "deleted": nil}
	update := bson.M{
		"$inc": bson.M{"tries": 1},
		"$set": bson.M{"ack": id(), "visible": time.Now().Add(c.Visibility)},
	}
	after := options.After
	res := msg.coll.FindOneAndUpdate(context.Background(), query, update, &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	})
	err := res.Err()
	if err != nil {
		return nil, false
	}
	message := new(QueueMessage)
	err = res.Decode(message)
	if err != nil {
		return nil, false
	}
	if c.MaxTries > 0 && message.Tries > int(c.MaxTries) {
		msg.coll.UpdateOne(
			context.Background(),
			bson.M{"ack": message.Ack},
			bson.M{"$set": bson.M{"dead": true, "deleted": time.Now()}},
		)
		return nil, false
	}
	return message, true
}

// Watch
func (msg *MessageNode) Watch(channel string, interval time.Duration) chan *QueueMessage {
	c := make(chan *QueueMessage)
	go func() {
		defer close(c)
		for {
			if q, ok := msg.Get(channel); ok {
				c <- q
			}
			time.Sleep(interval)
		}
	}()
	return c
}

type ackInfo struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Dead    bool               `bson:"dead"`
	Visible time.Time          `bson:"visible"`
	Deleted *time.Time         `bson:"deleted"`
}

// Ack
func (msg *MessageNode) Ack(ack string) bool {
	item := new(ackInfo)
	err := msg.coll.FindOne(context.Background(), bson.M{"ack": ack}).Decode(item)
	if err != nil {
		return false
	}
	if item.Visible.After(time.Now()) && !item.Dead && item.Deleted == nil {
		res, _ := msg.coll.UpdateByID(context.Background(), item.ID, bson.M{
			"$set": bson.M{"deleted": time.Now()},
		})
		return res.ModifiedCount == 1
	}
	return false
}

// Ping
func (msg *MessageNode) Ping(channel, ack string) bool {
	c, ok := msg.channel.Get(channel)
	if !ok {
		return false
	}
	query := bson.M{"ack": ack, "visible": bson.M{"$gt": time.Now()}, "dead": false, "deleted": nil}
	update := bson.M{
		"$set": bson.M{"visible": time.Now().Add(c.Visibility)},
	}
	res, err := msg.coll.UpdateOne(context.Background(), query, update)
	if err != nil {
		return false
	}
	return res.MatchedCount == 1
}

// Total
func (msg *MessageNode) Total(channel ...string) (int64, error) {
	if len(channel) >= 1 {
		return msg.coll.CountDocuments(context.Background(), bson.M{"channel": channel[0]})
	}
	return msg.coll.CountDocuments(context.Background(), bson.M{})
}

// Size
func (msg *MessageNode) Size(channel ...string) (int64, error) {
	query := bson.M{"visible": bson.M{"$lte": time.Now()}, "dead": false, "deleted": nil}
	if len(channel) >= 1 {
		query["channel"] = channel[0]
	}
	return msg.coll.CountDocuments(context.Background(), query)
}

// InFlight
func (msg *MessageNode) InFlight(channel ...string) (int64, error) {
	query := bson.M{"visible": bson.M{"$gt": time.Now()}, "dead": false, "deleted": nil}
	if len(channel) >= 1 {
		query["channel"] = channel[0]
	}
	return msg.coll.CountDocuments(context.Background(), query)
}

// Done
func (msg *MessageNode) Done(channel ...string) (int64, error) {
	query := bson.M{"deleted": bson.M{"$exists": true}}
	if len(channel) >= 1 {
		query["channel"] = channel[0]
	}
	return msg.coll.CountDocuments(context.Background(), query)
}

// Dead
func (msg *MessageNode) Dead(channel ...string) (int64, error) {
	query := bson.M{"dead": true}
	if len(channel) >= 1 {
		query["channel"] = channel[0]
	}
	return msg.coll.CountDocuments(context.Background(), query)
}

// MessageInt32
func (item *QueueMessage) MessageInt32() (int32, bool) {
	v, ok := item.Message.(int32)
	return v, ok
}

// MessageInt64
func (item *QueueMessage) MessageInt64() (int64, bool) {
	v, ok := item.Message.(int64)
	return v, ok
}

// MessageFloat64
func (item *QueueMessage) MessageFloat64() (float64, bool) {
	v, ok := item.Message.(float64)
	return v, ok
}

// MessageString
func (item *QueueMessage) MessageString() (string, bool) {
	v, ok := item.Message.(string)
	return v, ok
}

// MessageBool
func (item *QueueMessage) MessageBool() (bool, bool) {
	v, ok := item.Message.(bool)
	return v, ok
}

// MessageDecode
func (item *QueueMessage) MessageDecode(doc any) bool {
	switch item.Message.(type) {
	case primitive.D:
		return decode(item.Message, doc)
	case primitive.A:
		return decodeSlice(item.Message, doc)
	}
	return false
}

func decode(data any, doc any) bool {
	m, err := bson.Marshal(data)
	if err != nil {
		return false
	}
	if err = bson.Unmarshal(m, doc); err != nil {
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