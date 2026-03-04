package client

import (
	"path/filepath"
	"testing"

	nacoslogger "github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	internalconfig "nacos-cli/internal/config"
)

func TestBuildClientParam_DisableSnapshotAndCacheLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	param, err := BuildClientParam(internalconfig.Runtime{
		ServerAddr: "127.0.0.1:8848",
		Namespace:  "public",
	})
	if err != nil {
		t.Fatalf("BuildClientParam() error = %v", err)
	}
	if param.ClientConfig == nil {
		t.Fatalf("ClientConfig is nil")
	}
	if !param.ClientConfig.DisableUseSnapShot {
		t.Fatalf("DisableUseSnapShot = %v", param.ClientConfig.DisableUseSnapShot)
	}
	if !param.ClientConfig.NotLoadCacheAtStart {
		t.Fatalf("NotLoadCacheAtStart = %v", param.ClientConfig.NotLoadCacheAtStart)
	}
	if param.ClientConfig.CacheDir != filepath.Join(home, ".config", "nacos-cli", "cache") {
		t.Fatalf("CacheDir = %s", param.ClientConfig.CacheDir)
	}
}

func TestBuildClientParam_LogDisabledByDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	nacoslogger.SetLogger(nil)

	_, err := BuildClientParam(internalconfig.Runtime{
		ServerAddr: "127.0.0.1:8848",
		Namespace:  "public",
	})
	if err != nil {
		t.Fatalf("BuildClientParam() error = %v", err)
	}

	if _, ok := nacoslogger.GetLogger().(noopLogger); !ok {
		t.Fatalf("logger is not noop logger")
	}
}

func TestBuildClientParam_DevEnablesDebugLogDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	param, err := BuildClientParam(internalconfig.Runtime{
		ServerAddr: "127.0.0.1:8848",
		Namespace:  "public",
		Dev:        true,
	})
	if err != nil {
		t.Fatalf("BuildClientParam() error = %v", err)
	}
	if param.ClientConfig.LogLevel != "debug" {
		t.Fatalf("LogLevel = %s", param.ClientConfig.LogLevel)
	}
	if param.ClientConfig.LogDir != filepath.Join(home, ".config", "nacos-cli", "log") {
		t.Fatalf("LogDir = %s", param.ClientConfig.LogDir)
	}
}

func TestParseAddress_WithHTTPPrefix(t *testing.T) {
	host, port, err := parseAddress("http://10.0.16.100:8848")
	if err != nil {
		t.Fatalf("parseAddress() error = %v", err)
	}
	if host != "10.0.16.100" || port != 8848 {
		t.Fatalf("unexpected host/port: %s:%d", host, port)
	}
}

func TestParseAddress_WithHTTPSPrefixNoPort(t *testing.T) {
	host, port, err := parseAddress("https://10.0.16.100")
	if err != nil {
		t.Fatalf("parseAddress() error = %v", err)
	}
	if host != "10.0.16.100" || port != 8848 {
		t.Fatalf("unexpected host/port: %s:%d", host, port)
	}
}
