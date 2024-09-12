package mongo

import (
	"context"
	"log"
	"time"

	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

func Init(config Config) *mongodb.Database {
	clientOptions := options.Client().ApplyURI(config.URI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	// 选择数据库
	return client.Database(config.Database)
}
