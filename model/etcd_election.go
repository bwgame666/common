package model

import (
	"context"
	"errors"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
	"time"
)

type MasterEventType int

const (
	MasterAdded MasterEventType = iota
	MasterDeleted
	MasterModified
	MasterError
)

type MasterEvent struct {
	Type   MasterEventType
	Master string
	Error  error
}

type EtcdElectionClient struct {
	client        *clientv3.Client
	ctx           context.Context
	cancelContext context.CancelFunc
	session       *concurrency.Session
	election      *concurrency.Election
	eventChan     chan MasterEvent
	stopChan      chan struct{}
	key           string
	value         string
	callback      func(MasterEvent)
}

func NewEtcdElectionClient(client *clientv3.Client, key string, value string, cb func(MasterEvent)) (*EtcdElectionClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 15秒超时时间，得在15秒内续期
	session, err := concurrency.NewSession(client, concurrency.WithContext(ctx), concurrency.WithTTL(15))
	if err != nil {
		fmt.Println("Couldn't create etcd session: ", err, key)
		cancel()
		return nil, err
	}

	election := concurrency.NewElection(session, key)

	return &EtcdElectionClient{
		client:        client,
		ctx:           ctx,
		cancelContext: cancel,
		session:       session,
		election:      election,
		eventChan:     make(chan MasterEvent, 2),
		stopChan:      make(chan struct{}),
		key:           key,
		value:         value,
		callback:      cb,
	}, nil
}

func (e *EtcdElectionClient) Campaign() error {
	for {
		// 参与选举，尝试成为领导者
		ctx, cancelFunc := context.WithCancel(e.ctx)
		defer cancelFunc()
		if err := e.election.Campaign(ctx, e.value); err != nil {
			e.eventChan <- MasterEvent{Type: MasterError, Error: err}
			return err
		}
		//fmt.Println("Observe: ", e.key, e.value)
	observeEnd:
		for {
			select {
			case res := <-e.election.Observe(ctx):
				if len(res.Kvs) > 0 {
					if string(res.Kvs[0].Value) == e.value {
						fmt.Println("[Election] current node is be elected master: ", e.key, e.value)
						e.eventChan <- MasterEvent{Type: MasterAdded, Master: string(res.Kvs[0].Value)}
						break observeEnd
					} else {
					}
				}
			case <-e.session.Done():
				// 会话过期后尝试重新创建会话
				session, err := concurrency.NewSession(e.client, concurrency.WithContext(e.ctx), concurrency.WithTTL(15))
				if err != nil {
					fmt.Println("[Election] Failed to recreate session:", err)
					time.Sleep(5 * time.Second) // 等待后重试
					continue                    // 继续外层循环，重新参选
				}
				e.session = session
				e.election = concurrency.NewElection(session, e.key)
				continue
			}
		}
		//fmt.Println("select master: ", e.key, e.value)
		// if select master
	masterEnd:
		for {
			select {
			case <-e.ctx.Done():
				e.eventChan <- MasterEvent{Type: MasterError, Error: errors.New("elect: ctx done")}
				e.Resign()
				break masterEnd
			case <-e.session.Done():
				e.eventChan <- MasterEvent{Type: MasterError, Error: errors.New("elect: session expired")}
				//return errors.New("elect: session expired")
				break masterEnd
			}
		}
		// 等待1分钟后再重新参与选举
		time.Sleep(60 * time.Second)
		// 重新建立会话
		fmt.Println("[Election] renew session")
		session, err := concurrency.NewSession(e.client, concurrency.WithContext(e.ctx), concurrency.WithTTL(15))
		if err != nil {
			fmt.Println("Couldn't create etcd session: ", err, e.key)
		} else {
			e.session = session
			e.election = concurrency.NewElection(session, e.key)
		}
	}
}

func (e *EtcdElectionClient) Resign() error {
	// 释放领导权
	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()

	if err := e.election.Resign(ctx); err != nil {
		fmt.Println("[Election] Failed to resign", zap.Error(err))
		return err
	}
	fmt.Println("[Election] Resigned from leadership")
	return nil
}

func (e *EtcdElectionClient) Close() {
	_ = e.Resign()
	e.cancelContext()
	_ = e.session.Close()
	close(e.eventChan)
}

func (e *EtcdElectionClient) EventsChan() <-chan MasterEvent {
	return e.eventChan
}

func (e *EtcdElectionClient) Start() error {
	fmt.Println("EtcdElectionClient start: ", e.key, e.value)
	// 开启新goroutine, 监听选举情况
	go e.Campaign()
	// 开启新goroutine, 处理信号
	go e.onEvent()

	return nil
}

func (e *EtcdElectionClient) Stop() {
	fmt.Println("EtcdElectionClient stop: ", e.key, e.value)
	e.stopChan <- struct{}{}
	e.Close()
}

func (e *EtcdElectionClient) onEvent() {
	for {
		select {
		case event := <-e.EventsChan():
			e.callback(event)
		case <-e.stopChan:
			fmt.Println("[Election] Elect onEvent stopped: ", e.key, e.value)
			return
		}
	}
}
