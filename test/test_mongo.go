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
		dbId, err := mongoClient.AddOne(&TestI{
			Test: 6,
		})
		if err != nil {
			return nil, err
		}
		var data TestI
		err2 := mongoClient.GetOne(dbId, &data)
		fmt.Println(data)
		if err2 != nil {
			fmt.Println("get error: ", err2)
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

	s := mongoClient.StartSession()
	defer mongoClient.EndSession(s)

	result, err := mongoClient.StatTransaction(s, callback)
	if err != nil {
		fmt.Println(err)
		fmt.Println(result)
	}
	fmt.Println(result)
}

func testReadWrite() {
	mongoClient, err := model.NewMongoClient("bac", "test1")
	if err != nil {
		fmt.Println("new mongo client failed: ")
	}
	_, err = mongoClient.AddOne(&TestI{
		Test: 6,
	})
	var data []TestI
	_ = mongoClient.Query(map[string]interface{}{}, 0, 100, &data)
	fmt.Println(data)
}

func test() {
	mongoClient, err := model.NewMongoClient("bac", "test1")
	if err != nil {
		fmt.Println("new mongo client failed: ")
	}
	ObjId, err := mongoClient.AddOne(&TestI{
		Test: 5,
	})
	if err != nil {
		fmt.Println("5 failed: ", err)
	}
	err = mongoClient.UpdateOne(ObjId, &TestI{
		Test: 6,
	})
	fmt.Println(err)
}

func main() {
	qClient := model.InitMongoConnection("mongodb://35.197.129.138:27020,35.197.129.138:27018,35.197.129.138:27019/?replicaSet=rs0",
		"bac", "bac@123", "bac")
	if qClient == nil {
		fmt.Println("connect failed")
		return
	}
	//test()
	//testTransaction()
	testReadWrite()
}
