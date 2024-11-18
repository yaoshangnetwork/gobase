package mqueue

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MQueue struct {
	opts    *QueueOpts
	Message MessageNode
}

type Mode string

const (
	DebugMode         Mode = "debug"
	ReleaseMode       Mode = "release"
	DefaultVisibility      = time.Minute * 5
)

type Channel struct {
	Name       string
	Collection string
}

type QueueOpts struct {
	Mode       Mode
	DB         DBConfig
	Visibility time.Duration
}

func NewMQueue(opts QueueOpts) *MQueue {
	if opts.Mode == "" {
		opts.Mode = ReleaseMode
	}
	database := connect(opts.DB)

	mq := &MQueue{
		opts: &opts,
		Message: MessageNode{
			mode:       opts.Mode,
			coll:       database.Collection(opts.DB.Collection),
			visibility: opts.Visibility,
		},
	}

	mq.createIndexes()
	return mq
}

// createIndexes
func (mq *MQueue) createIndexes() {
	// mq.Message.coll.Indexes().DropAll(context.Background())

	_, err := mq.Message.coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "visible", Value: 1},
			{Key: "dead", Value: 1},
			{Key: "deleted", Value: 1},
		},
	})
	if err != nil {
		panic(err)
	}

	// ack 唯一索引
	_, err = mq.Message.coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "ack", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		panic(err)
	}

	// 已完成的 queue message 保留 7 天数据
	_, err = mq.Message.coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "deleted", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(3600 * 24 * 7),
	})
	if err != nil {
		panic(err)
	}
}
