package main

import (
	"context"
	"fmt"
	"github.com/bwgame666/common/model"
)

type TestI struct {
	Test int32 `bson:"test"`
}

func testTransaction() {
	mongoClient, err := model.NewMongoClient("bac", "test1")
	if err != nil {
		fmt.Println("new mongo client failed: ")
	}
	mongoClient2, err := model.NewMongoClient("bac", "test2")
	if err != nil {
		fmt.Println("new mongo client failed: ")
	}

	callback := func(sessCtx context.Context) (interface{}, error) {
		mongoClient.SetContext(sessCtx)
		mongoClient2.SetContext(sessCtx)
		_, err = mongoClient.AddOne(&TestI{
			Test: 6,
		})
		if err != nil {
			return nil, err
		}
		_, err = mongoClient2.AddOne(&TestI{
			Test: 6,
		})
		if err != nil {
			return nil, err
		}
		return "success", nil
	}

	mongoClient.StartSession()
	defer mongoClient.EndSession()

	result, err := mongoClient.StatTransaction(callback)
	if err != nil {
		fmt.Println(err)
		fmt.Println(result)
	}
	fmt.Println(result)
}

func test() {
	mongoClient, err := model.NewMongoClient("bac", "test1")
	if err != nil {
		fmt.Println("new mongo client failed: ")
	}
	_, err = mongoClient.AddOne(&TestI{
		Test: 3,
	})
	if err != nil {
		fmt.Println("3 failed: ", err)
	}
	_, err = mongoClient.AddOne(&TestI{
		Test: 5,
	})
	if err != nil {
		fmt.Println("5 failed: ", err)
	}
}

func main() {
	qClient := model.InitMongoConnection("mongodb://127.0.0.1:27020/?replicaSet=rs0",
		"user", "pass", "db")
	if qClient == nil {
		fmt.Println("connect failed")
		return
	}
	testTransaction()
}
