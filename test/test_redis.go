package main

import (
	"fmt"
	"github.com/bwgame666/common/model"
)

func main() {
	_ = model.InitRedis("35.197.129.138:6379", "admin123", 0)
	redisClient, _ := model.NewRedisClient()
	data := map[string]interface{}{
		"name": "Hello",
		"age":  int32(60),
	}
	_ = redisClient.HSet("testH", data)
	result, _ := redisClient.HGetALL("testH")
	fmt.Printf("type: %T --", result["age"])
	fmt.Println(result["name"], result["age"])

	data2 := 30
	_ = redisClient.Set("test", data2)

	r2, _ := redisClient.Get("test")

	fmt.Printf("type: %T --", r2)
	fmt.Println(r2)

	_ = redisClient.HSetBy("testH", "age", 90)
	_ = redisClient.HIncrBy("testH", "age", 90)
	r, _ := redisClient.HGetBy("testH", "age")
	fmt.Printf("type: %T --", r)
	fmt.Println(r)

}
