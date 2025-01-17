package db

import (
	"context"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var client *mongo.Client

func MongoViper(ctx context.Context) *mongo.Client {
	if client == nil {
		mongoURI := viper.GetString("mongo-uri")
		var err error
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			zap.L().Panic("error creating MongoDB client", zap.Error(err))
		}
	}
	return client
}
