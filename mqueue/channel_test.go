package mqueue_test

import (
	"testing"

	"github.com/yaoshangnetwork/gobase/mqueue"
)

func getMQ() *mqueue.MQueue {
	return mqueue.NewMQueue(mqueue.QueueOpts{
		Mode: mqueue.DebugMode,
		DB: mqueue.DBConfig{
			URI:               "mongodb://127.0.0.1:27017",
			Database:          "dp_test",
			MessageCollection: "queue_messages",
			ChannelCollection: "queue_channels",
		},
	})
}

func TestChannel(t *testing.T) {
	mq := getMQ()

	// clear
	mq.Channel.Clear()

	// create
	ok := mq.Channel.Create(&mqueue.Channel{Name: "123"})
	if ok != true {
		t.Error("err")
	}
	ok = mq.Channel.Create(&mqueue.Channel{Name: "123"})
	if ok == true {
		t.Error("err")
	}

	// get
	c, ok := mq.Channel.Get("123")
	if !ok {
		t.Error("err")
	}
	if c.Name != "123" {
		t.Error("err")
	}

	// update
	ok = mq.Channel.Update(&mqueue.Channel{Name: "123", MaxTries: 2})
	if !ok {
		t.Error("err")
	}

	// get
	c, ok = mq.Channel.Get("123")
	if !ok {
		t.Error("err")
	}
	if c.MaxTries != 2 {
		t.Error("err")
	}

	// delete
	ok = mq.Channel.Delete("123")
	if !ok {
		t.Error("err")
	}

	// get
	_, ok = mq.Channel.Get("123")
	if ok {
		t.Error("err")
	}

}
