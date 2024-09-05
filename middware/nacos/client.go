package nacos

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"log"
)

type Client struct {
	client config_client.IConfigClient
}

// InitConfigClient 初始化 Nacos 配置客户端
func InitConfigClient(ip string, port uint64, namespaceID string) *Client {
	// Nacos服务器地址配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      ip,
			ContextPath: "/nacos",
			Port:        port,
			Scheme:      "http",
		},
	}
	// Nacos客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         namespaceID, // 命名空间ID
		TimeoutMs:           5000,        // 请求超时时间
		NotLoadCacheAtStart: false,       // 是否在启动时加载缓存
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		Username:            "nacos",
		Password:            "nacos",
		LogLevel:            "debug",
	}

	// 创建 Nacos 配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Nacos config client: %v", err))
	}

	return &Client{client: configClient}
}

// GetConfig 从 Nacos 获取配置信息
func (n *Client) GetConfig(dataID, group string) (string, error) {
	return n.client.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
}

// ListenConfig 监听 Nacos 配置变化并执行相应操作
func (n *Client) ListenConfig(dataID, group string, onChange func(string)) error {
	return n.client.ListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
		OnChange: func(namespace, group, dataID, data string) {
			// 当配置发生变化时，执行注册的回调函数
			log.Printf("Config changed - Namespace: %s, Group: %s, DataId: %s", namespace, group, dataID)
			onChange(data)
		},
	})
}

// ParseJSONConfig 解析JSON配置
func (n *Client) ParseJSONConfig(jsonStr string, target interface{}) error {
	err := json.Unmarshal([]byte(jsonStr), target)
	if err != nil {
		return fmt.Errorf("failed to parse JSON config: %v", err)
	}
	return nil
}
