package client

import (
	"testing"

	internalconfig "nacos-cli/internal/config"
)

func TestBuildClientParam_DisableSnapshotAndCacheLoad(t *testing.T) {
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
