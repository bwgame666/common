package main

import (
	"fmt"
	"github.com/bwgame666/common/model"
	"time"
)

var (
	master int = 0
)

func CallBack1(m model.MasterEvent) {
	master = 1
	fmt.Println("callback1: ", m)
}

func CallBack2(m model.MasterEvent) {
	master = 2
	fmt.Println("callback2: ", m)
}

func CallBack3(m model.MasterEvent) {
	master = 3
	fmt.Println("callback3: ", m)
}

func main() {
	etcC, err := model.NewEtcdClient([]string{"http://35.197.129.138:2379"}, "root", "admin@2023!")
	if err != nil {
		fmt.Println("connect etcd failed: ", err)
		return
	}
	elect1, _ := model.NewEtcdElectionClient(etcC.GetClient(), "test", "thread01", CallBack1)
	elect2, _ := model.NewEtcdElectionClient(etcC.GetClient(), "test", "thread02", CallBack2)
	elect3, _ := model.NewEtcdElectionClient(etcC.GetClient(), "test", "thread03", CallBack3)

	go elect1.Start()
	go elect2.Start()
	go elect3.Start()

	fmt.Println("hhhhhh")

	for {
		time.Sleep(30 * time.Second)
		if master == 1 {
			elect1.Stop()
		}
		if master == 2 {
			elect2.Stop()
		}
		if master == 3 {
			elect3.Stop()
		}
		master = 0

	}
}
