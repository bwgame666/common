package model

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/client/v3"
	"math"
	"time"
)

type EtcdClient struct {
	client *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEtcdClient(endpoints []string, userName string, password string) (*EtcdClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	config := clientv3.Config{
		Endpoints:          endpoints,
		MaxCallRecvMsgSize: math.MaxInt32,
		Context:            ctx,
		DialTimeout:        5 * time.Second,
	}
	if userName != "" {
		config.Username = userName
	}
	if password != "" {
		config.Password = password
	}
	client, err := clientv3.New(config)
	if err != nil {
		cancel()
		return nil, err
	}
	return &EtcdClient{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (that *EtcdClient) SetConfig(path string, doc interface{}) error {
	ctx := context.Background()
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = that.client.Put(ctx, path, string(data))
	return err
}

func (that *EtcdClient) GetConfig(path string, result interface{}) error {
	ctx := context.Background()
	resp, err := that.client.Get(ctx, path)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return nil
	}
	err = json.Unmarshal(resp.Kvs[0].Value, result)
	return err
}

func (that *EtcdClient) Close() error {
	if that.client == nil {
		return nil
	}
	that.cancel()

	return that.client.Close()
}
