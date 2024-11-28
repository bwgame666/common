package mq

import (
	"github.com/bwgame666/common/libs"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
)

type MqttClient struct {
	Client mqtt.Client
}

func NewMqttClient(broker string, username string, passwd string) *MqttClient {
	/*
		// 添加多个 broker
		    for _, broker := range brokers {
		        opts.AddBroker(broker)
		    }
	*/
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(libs.RandStr(16)).
		SetUsername(username).
		SetPassword(passwd).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(1 * time.Second)

	// 创建并启动客户端
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
		return nil
	}
	return &MqttClient{
		Client: client,
	}
}

func (m *MqttClient) Public(topic string, message string) {
	token := m.Client.Publish(topic, 0, false, message)
	token.Wait()
}

func (m *MqttClient) Close() {
	m.Client.Disconnect(250)
}
