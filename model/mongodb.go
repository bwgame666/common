package model

import (
	"context"
	"fmt"
	"github.com/qiniu/qmgo"
	qnOpts "github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"sync"
)

var (
	mongoClient *qmgo.Client
	mongoOnce   sync.Once
)

type MongoClient struct {
	client     *qmgo.Client
	database   *qmgo.Database
	collection *qmgo.Collection
	sess       *qmgo.Session
	ctx        context.Context
}

func InitMongoConnection(url string, username string, passwd string, dbname string) *qmgo.Client {
	var (
		timeout     int64  = 2000
		maxPoolSize uint64 = 100
		minPoolSize uint64 = 0
	)
	var ctx = context.Background()
	mongoOnce.Do(func() {
		clientOptions := &options.ClientOptions{}
		// 设置认证信息
		credential := options.Credential{
			Username:   username,
			Password:   passwd,
			AuthSource: dbname,
		}
		clientOptions.SetAuth(credential)
		opts := qnOpts.ClientOptions{
			ClientOptions: clientOptions,
		}

		cfg := qmgo.Config{
			Uri:              url,
			ConnectTimeoutMS: &timeout,
			MaxPoolSize:      &maxPoolSize,
			MinPoolSize:      &minPoolSize,
			ReadPreference:   &qmgo.ReadPref{Mode: readpref.SecondaryMode, MaxStalenessMS: 100 * 1000},
		}

		cli, err := qmgo.NewClient(ctx, &cfg, opts)
		if err != nil {
			fmt.Println("initMongoClient failed: ", err.Error())
			return
		}
		err = cli.Ping(5)
		if err != nil {
			fmt.Println("MongoClient ping failed: ", err.Error())
			return
		}
		mongoClient = cli
	})
	return mongoClient
}

func NewMongoClient(dbName, collectionName string) (*MongoClient, error) {
	// 获取数据库和集合
	database := mongoClient.Database(dbName)
	collection := database.Collection(collectionName)

	return &MongoClient{
		client:     mongoClient,
		database:   database,
		collection: collection,
		ctx:        context.TODO(),
	}, nil
}

func (that *MongoClient) StartSession() *qmgo.Session {
	opts := &qnOpts.SessionOptions{
		SessionOptions: &options.SessionOptions{
			DefaultReadPreference: readpref.Primary(), // 设置为主节点读取偏好, read preference in a transaction must be primary
		},
	}
	s, err := that.client.Session(opts)
	if err != nil {
		fmt.Println("start mongo session failed: ", err)
	}
	that.sess = s
	return that.sess
}

func (that *MongoClient) EndSession() {
	that.sess.EndSession(context.Background())
	that.sess = nil
}

func (that *MongoClient) StatTransaction(cb func(sessCtx context.Context) (interface{}, error)) (interface{}, error) {
	result, err := that.sess.StartTransaction(context.Background(), cb)
	if err != nil {
		fmt.Println("StartTransaction failed, err=", err)
		return result, err
	}
	return result, nil
}

func (that *MongoClient) SetContext(ctx context.Context) {
	that.ctx = ctx
}

func (that *MongoClient) GetContext() context.Context {
	return that.ctx
}

func (that *MongoClient) GetCollection() *qmgo.Collection {
	return that.collection
}

func (that *MongoClient) AddOne(doc interface{}) (string, error) {
	result, err := that.collection.InsertOne(that.ctx, doc)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (that *MongoClient) GetOne(id string, result interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := map[string]interface{}{"_id": objectID}
	err = that.collection.Find(that.ctx, filter).One(result)
	if err != nil {
		return err
	}
	return nil
}

func (that *MongoClient) UpdateOne(id string, doc interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := map[string]interface{}{"_id": objectID}
	update := map[string]interface{}{"$set": doc}
	err = that.collection.UpdateOne(that.ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (that *MongoClient) DeleteOne(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	var result interface{}
	err = that.GetOne(id, result)
	if err != nil {
		return err
	}
	err = that.collection.RemoveId(that.ctx, objectID)
	if err != nil {
		return err
	}
	return nil
}

func (that *MongoClient) Query(filter interface{}, start int64, count int64, result interface{}) error {
	err := that.collection.Find(that.ctx, filter).Skip(start).Limit(count).All(result)
	if err != nil {
		return err
	}
	return nil
}
