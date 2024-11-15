package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml"
	"go.etcd.io/etcd/client/v3"
	"math"
	"time"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 8 * time.Second
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
		DialTimeout:        dialTimeout,
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

/*
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
*/

func isDigit(s string) bool {

	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (that *EtcdClient) GetClient() *clientv3.Client {
	return that.client
}

func (that *EtcdClient) Parse(key string, v interface{}) error {

	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	kv := clientv3.NewKV(that.client)
	gr, _ := kv.Get(ctx, key)
	if gr == nil || len(gr.Kvs) == 0 {
		return fmt.Errorf("no more '%s'", key)
	}

	return json.Unmarshal(gr.Kvs[0].Value, v)
}

func (that *EtcdClient) ParseTomlStruct(key string, v interface{}) error {

	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	kv := clientv3.NewKV(that.client)
	gr, _ := kv.Get(ctx, key)
	if gr == nil || len(gr.Kvs) == 0 {
		fmt.Println("no more: ", key)
		return fmt.Errorf("no more '%s'", key)
	}

	return toml.Unmarshal(gr.Kvs[0].Value, v)
}

func (that *EtcdClient) ParseToml(key string, filter bool) (map[string]map[string]interface{}, error) {

	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	kv := clientv3.NewKV(that.client)
	gr, _ := kv.Get(ctx, key)
	if gr == nil || len(gr.Kvs) == 0 {
		return nil, fmt.Errorf("no more '%s'", key)
	}

	recs := map[string]map[string]interface{}{}
	config, err := toml.LoadBytes(gr.Kvs[0].Value)
	if err != nil {
		return recs, err
	}

	keys := config.Keys()
	for _, val := range keys {

		if filter && isDigit(val) {

			tree := config.Get(val).(*toml.Tree)
			recs[val] = tree.ToMap()
		} else {
			tree := config.Get(val).(*toml.Tree)
			recs[val] = tree.ToMap()
		}
	}

	return recs, nil
}

func (that *EtcdClient) Close() error {
	if that.client == nil {
		return nil
	}
	that.cancel()

	return that.client.Close()
}
