package model

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	client *clientv3.Client
}

func NewEtcdClient(endpoints []string) (*EtcdClient, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdClient{
		client: client,
	}, nil
}

func (that *EtcdClient) SetConfig(ctx context.Context, path string, doc interface{}) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = that.client.Put(ctx, path, string(data))
	return err
}

func (that *EtcdClient) GetConfig(ctx context.Context, path string) (interface{}, error) {
	resp, err := that.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	var result interface{}
	err = json.Unmarshal(resp.Kvs[0].Value, &result)
	return result, err
}
