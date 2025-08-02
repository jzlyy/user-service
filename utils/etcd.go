package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"user-service/config"

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

func (e *EtcdClient) RegisterService(serviceName string, port int) error {
	cfg := config.LoadConfig()
	ttl := cfg.EtcdTTL // 从配置获取TTL

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

	// 启动健康检查协程
	go e.healthCheck(serviceName, port)

	return nil
}

func (e *EtcdClient) healthCheck(serviceName string, port int) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 使用localhost可能有问题，改为相对路径
		url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
		resp, err := http.Get(url)

		// 处理可能的错误
		if err != nil {
			log.Printf("Health check failed: %v", err)
			continue
		}

		// 确保关闭响应体
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Health check failed: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			log.Printf("Health check failed with status: %d", resp.StatusCode)
			continue
		}

		// 刷新租约
		if _, err := e.client.KeepAliveOnce(context.Background(), e.lease); err != nil {
			log.Printf("Failed to refresh lease: %v", err)
		}
	}
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
