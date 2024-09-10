package main

import (
	"fmt"
	"github.com/bwgame666/common/mq"
)

func main() {
	var broker string = "tcp://35.197.129.138:21883"
	natClient := mq.NewMqttClient(broker, "admin", "bac123")
	if natClient == nil {
		fmt.Println("NewMqttClient failed: ")
		return
	}

	natClient.Public("bac/test", "abc")

}
