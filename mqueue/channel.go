package mqueue

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChannelNode struct {
	coll  *mongo.Collection
	mode  Mode
	cache []*Channel
}

type Channel struct {
	Name       string        `bson:"name"`
	Visibility time.Duration `bson:"visibility"`
	MaxTries   int64         `bson:"max_tries"`
}

type IChannelNode interface {
	Clear() bool // dev mode only
	Create(channel *Channel) bool
	Update(channel *Channel) bool
	Delete(name string) bool
	Get(name string) (*Channel, bool)
	All() ([]*Channel, bool)
	Sync()
}

var _ IChannelNode = (*ChannelNode)(nil)

// Clear
func (c *ChannelNode) Clear() bool {
	if c.mode != DebugMode {
		log.Print("the \"Clean\" method can only be used in debug mode")
		return false
	}
	_, err := c.coll.DeleteMany(context.Background(), bson.M{})
	if err == nil {
		c.cache = nil
	}
	return err == nil
}

// Create
func (c *ChannelNode) Create(channel *Channel) bool {
	if channel == nil {
		return false
	}
	if channel.Name == "" {
		return false
	}
	if channel.Visibility == 0 {
		channel.Visibility = DefaultVisibility
	}
	_, err := c.coll.InsertOne(context.Background(), channel)
	if err == nil {
		c.updateCache()
	}
	return err == nil
}

// Update
func (c *ChannelNode) Update(channel *Channel) bool {
	if channel == nil {
		return false
	}
	if channel.Name == "" {
		return false
	}
	if channel.Visibility == 0 {
		channel.Visibility = DefaultVisibility
	}
	_, err := c.coll.UpdateOne(
		context.Background(),
		bson.M{"name": channel.Name},
		bson.M{"$set": channel},
	)
	if err == nil {
		c.updateCache()
	}
	return err == nil
}

// Delete
func (c *ChannelNode) Delete(name string) bool {
	_, err := c.coll.DeleteOne(context.Background(), bson.M{"name": name})
	if err == nil {
		c.updateCache()
	}
	return err == nil
}

// Get
func (c *ChannelNode) Get(name string) (*Channel, bool) {
	for _, channel := range c.cache {
		if channel.Name == name {
			return channel, true
		}
	}
	return nil, false
}

// All
func (c *ChannelNode) All() ([]*Channel, bool) {
	if c.cache == nil || len(c.cache) == 0 {
		return nil, false
	}
	return c.cache, true
}

// Sync
func (c *ChannelNode) Sync() {
	c.updateCache()
}

// updateCache
func (c *ChannelNode) updateCache() {
	channels := make([]*Channel, 0)
	cur, err := c.coll.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	defer cur.Close(context.Background())
	err = cur.All(context.Background(), &channels)
	if err != nil {
		return
	}
	c.cache = channels
}
