package main

import (
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/bwgame666/common/mq"
	"github.com/nats-io/nats.go"
)

func NatMsg(msg *nats.Msg) {
	fmt.Println("Received message: ", string(msg.Data))

}

func main() {
	var url string = "nats://admin:admin123@35.197.129.138:4222"
	natClient, err := mq.NewNatsClient(url)
	if err != nil {
		fmt.Println("connect nats failed: ", err)
		return
	}

	err = natClient.Subscript("bac_tower_notify_inst001", NatMsg)
	if err != nil {
		fmt.Println("Failed to create subscriber: ", err)
	}

	_ = libs.RunForever()
}
