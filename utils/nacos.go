package utils

import (
	"fmt"
	"log"
	"net"
	"user-service/config"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type NacosClient struct {
	client      naming_client.INamingClient
	cfg         *config.Config // 添加配置字段
	serviceName string
	ip          string
	port        uint64
}

func NewNacosClient(cfg *config.Config) (*NacosClient, error) {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      cfg.NacosAddresses,
			Port:        8848,
			Scheme:      "http",
			ContextPath: "/nacos",
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.NacosNamespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
		Username:            cfg.NacosUsername,
		Password:            cfg.NacosPassword,
	}

	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, err
	}

	ip, err := getLocalIP()
	if err != nil {
		return nil, err
	}

	return &NacosClient{
		client:      namingClient,
		cfg:         cfg, // 保存配置
		serviceName: cfg.ServiceName,
		ip:          ip,
		port:        uint64(cfg.ServicePort),
	}, nil
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer func(conn net.Conn) {
			err := conn.Close()
			if err != nil {

			}
		}(conn)
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String(), nil
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no valid local IP found")
}

func (n *NacosClient) RegisterService() error {
	instance := vo.RegisterInstanceParam{
		Ip:          n.ip,
		Port:        n.port,
		ServiceName: n.serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"secure": "false"},
		ClusterName: n.cfg.NacosCluster,
		GroupName:   n.cfg.NacosGroup,
	}

	success, err := n.client.RegisterInstance(instance)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("failed to register service")
	}
	log.Printf("Registered service %s at %s:%d", n.serviceName, n.ip, n.port)
	return nil
}

func (n *NacosClient) DeregisterService() error {
	instance := vo.DeregisterInstanceParam{
		Ip:          n.ip,
		Port:        n.port,
		ServiceName: n.serviceName,
		Cluster:     n.cfg.NacosCluster,
		GroupName:   n.cfg.NacosGroup,
		Ephemeral:   true,
	}
	success, err := n.client.DeregisterInstance(instance)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("failed to deregister service")
	}
	log.Println("Deregistered service")
	return nil
}
