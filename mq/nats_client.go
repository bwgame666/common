package mq

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	conn   *nats.Conn
	subMap map[string]*nats.Subscription
}

func NewNatsClient(url string) (*NatsClient, error) {
	natsConn, err := nats.Connect(url)
	if err != nil {
		fmt.Println("Failed to connect to NATS: ", url, err)
		return nil, err
	}
	return &NatsClient{
		conn:   natsConn,
		subMap: map[string]*nats.Subscription{},
	}, nil
}

func (nc *NatsClient) Public(topic string, msg []byte) error {
	if err := nc.conn.Publish(topic, msg); err != nil {
		fmt.Println("Failed to publish message: ", topic, err)
		return err
	}
	return nil
}

func (nc *NatsClient) Subscript(topic string, fn nats.MsgHandler) error {
	sub, err := nc.conn.Subscribe(topic, fn)
	if err != nil {
		fmt.Println("Failed to create subscriber: ", topic, err)
		return err
	}
	nc.subMap[topic] = sub
	return nil
}

func (nc *NatsClient) UnSubscript(topic string) error {
	err := nc.subMap[topic].Unsubscribe()
	if err != nil {
		fmt.Println("Failed to Unsubscribe: ", topic, err)
		return err
	}
	return nil
}
