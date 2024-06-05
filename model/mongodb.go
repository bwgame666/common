package model

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
)

var (
	mongoClient *mongo.Client
	mongoOnce   sync.Once
)

type MongoClient struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func NewMongoClient(uri, dbName, collectionName string) (*MongoClient, error) {

	mongoOnce.Do(func() {
		// 单例模式，创建 MongoDB 客户端, 进程复用一个连接
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			return
		}
		mongoClient = client
	})
	// 获取数据库和集合
	database := mongoClient.Database(dbName)
	collection := database.Collection(collectionName)

	return &MongoClient{
		client:     mongoClient,
		database:   database,
		collection: collection,
	}, nil
}

func (that *MongoClient) AddOne(doc interface{}) (string, error) {
	result, err := that.collection.InsertOne(context.TODO(), doc)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (that *MongoClient) GetOne(id string) (interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = that.collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (that *MongoClient) UpdateOne(id string, doc interface{}) (interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	_, err = that.collection.ReplaceOne(context.TODO(), bson.M{"_id": objectID}, doc)
	if err != nil {
		return nil, err
	}
	return that.GetOne(id)
}

func (that *MongoClient) DeleteOne(id string) (interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	result, err := that.GetOne(id)
	if err != nil {
		return nil, err
	}
	_, err = that.collection.DeleteOne(context.TODO(), bson.M{"_id": objectID})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (that *MongoClient) Query(filter interface{}) ([]interface{}, error) {
	cur, err := that.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			fmt.Println(err.Error())
		}
	}(cur, context.TODO())
	var results []interface{}
	for cur.Next(context.TODO()) {
		var result interface{}
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
