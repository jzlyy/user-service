package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	client *clientv3.Client
	lease  clientv3.LeaseID
}

func NewEtcdClient(address, username, password string) (*EtcdClient, error) {
	cfg := clientv3.Config{
		Endpoints:   []string{address},
		DialTimeout: 5 * time.Second,
		Username:    username,
		Password:    password,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &EtcdClient{client: client}, nil
}

func (e *EtcdClient) RegisterService(serviceName string, port int, ttl int) error {
	// 创建租约
	resp, err := e.client.Grant(context.Background(), int64(ttl))
	if err != nil {
		return err
	}

	// 服务唯一ID
	serviceID := fmt.Sprintf("%s-%d", serviceName, time.Now().UnixNano())
	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceID)
	value := fmt.Sprintf("localhost:%d", port)

	// 注册服务
	_, err = e.client.Put(
		context.Background(),
		key,
		value,
		clientv3.WithLease(resp.ID),
	)
	if err != nil {
		return err
	}

	e.lease = resp.ID

	// 保持租约
	go e.keepAlive()

	log.Printf("Registered in ETCD: %s -> %s", key, value)
	return nil
}

func (e *EtcdClient) keepAlive() {
	ch, err := e.client.KeepAlive(context.Background(), e.lease)
	if err != nil {
		log.Printf("Failed to keep lease alive: %v", err)
		return
	}

	for range ch {
		// 自动处理通道关闭
	}
	log.Println("ETCD keepalive channel closed")
}

func (e *EtcdClient) DeregisterService() error {
	if e.lease != 0 {
		_, err := e.client.Revoke(context.Background(), e.lease)
		return err
	}
	return nil
}

func (e *EtcdClient) Close() {
	if e.client != nil {
		err := e.client.Close()
		if err != nil {
			return
		}
	}
}
