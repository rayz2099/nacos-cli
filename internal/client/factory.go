package client

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	nacoslogger "github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	internalconfig "nacos-cli/internal/config"
)

type ConfigClient interface {
	GetConfig(param vo.ConfigParam) (string, error)
	PublishConfig(param vo.ConfigParam) (bool, error)
	DeleteConfig(param vo.ConfigParam) (bool, error)
	SearchConfig(param vo.SearchConfigParam) (*model.ConfigPage, error)
	CloseClient()
}

type NamingClient interface {
	RegisterInstance(param vo.RegisterInstanceParam) (bool, error)
	DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error)
	SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error)
	CloseClient()
}

func NewConfigClient(cfg internalconfig.Runtime) (ConfigClient, error) {
	param, err := BuildClientParam(cfg)
	if err != nil {
		return nil, err
	}

	client, err := clients.NewConfigClient(param)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewNamingClient(cfg internalconfig.Runtime) (NamingClient, error) {
	param, err := BuildClientParam(cfg)
	if err != nil {
		return nil, err
	}

	client, err := clients.NewNamingClient(param)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func BuildClientParam(cfg internalconfig.Runtime) (vo.NacosClientParam, error) {
	serverConfigs, err := parseServerConfigs(cfg.ServerAddr)
	if err != nil {
		return vo.NacosClientParam{}, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return vo.NacosClientParam{}, err
	}

	baseDir := filepath.Join(homeDir, ".config", "nacos-cli")
	cacheDir := filepath.Join(baseDir, "cache")
	clientOptions := []constant.ClientOption{
		constant.WithNamespaceId(cfg.Namespace),
		constant.WithUsername(cfg.Username),
		constant.WithPassword(cfg.Password),
		constant.WithDisableUseSnapShot(true),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithCacheDir(cacheDir),
	}
	if cfg.Dev {
		clientOptions = append(clientOptions,
			constant.WithLogDir(filepath.Join(baseDir, "log")),
			constant.WithLogLevel("debug"),
		)
		nacoslogger.SetLogger(nil)
	} else {
		nacoslogger.SetLogger(noopLogger{})
	}

	clientConfig := constant.NewClientConfig(clientOptions...)

	return vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	}, nil
}

func parseServerConfigs(raw string) ([]constant.ServerConfig, error) {
	parts := strings.Split(raw, ",")
	result := make([]constant.ServerConfig, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}

		host, port, err := parseAddress(item)
		if err != nil {
			return nil, err
		}
		result = append(result, *constant.NewServerConfig(host, port))
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("server-addr is required")
	}

	return result, nil
}

func parseAddress(addr string) (string, uint64, error) {
	value := strings.TrimSpace(addr)
	if value == "" {
		return "", 0, fmt.Errorf("invalid server address: %s", addr)
	}

	if strings.Contains(value, "://") {
		u, err := url.Parse(value)
		if err != nil {
			return "", 0, fmt.Errorf("invalid server address: %s", addr)
		}
		host := strings.TrimSpace(u.Hostname())
		if host == "" {
			return "", 0, fmt.Errorf("invalid server address: %s", addr)
		}
		port := u.Port()
		if port == "" {
			return host, 8848, nil
		}
		p, err := strconv.ParseUint(strings.TrimSpace(port), 10, 64)
		if err != nil {
			return "", 0, fmt.Errorf("invalid server port in address %s", addr)
		}
		return host, p, nil
	}

	lastColon := strings.LastIndex(value, ":")
	if lastColon == -1 {
		return value, 8848, nil
	}

	host := strings.TrimSpace(value[:lastColon])
	portRaw := strings.TrimSpace(value[lastColon+1:])
	if host == "" || portRaw == "" {
		return "", 0, fmt.Errorf("invalid server address: %s", addr)
	}

	port, err := strconv.ParseUint(portRaw, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid server port in address %s", addr)
	}

	return host, port, nil
}

type noopLogger struct{}

func (noopLogger) Info(...interface{}) {}

func (noopLogger) Warn(...interface{}) {}

func (noopLogger) Error(...interface{}) {}

func (noopLogger) Debug(...interface{}) {}

func (noopLogger) Infof(string, ...interface{}) {}

func (noopLogger) Warnf(string, ...interface{}) {}

func (noopLogger) Errorf(string, ...interface{}) {}

func (noopLogger) Debugf(string, ...interface{}) {}

func (noopLogger) Close() error { return nil }

var _ config_client.IConfigClient
var _ naming_client.INamingClient
