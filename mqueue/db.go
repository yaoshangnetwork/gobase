package mqueue

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBConfig struct {
	URI        string
	Database   string
	Collection string
}

func connect(config DBConfig) *mongo.Database {
	clientOptions := options.Client().ApplyURI(config.URI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	// 选择数据库
	return client.Database(config.Database)
}
