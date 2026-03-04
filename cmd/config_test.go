package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/cobra"
	internalclient "nacos-cli/internal/client"
	internalconfig "nacos-cli/internal/config"
)

type mockConfigClient struct {
	getConfigFn     func(param vo.ConfigParam) (string, error)
	publishConfigFn func(param vo.ConfigParam) (bool, error)
	deleteConfigFn  func(param vo.ConfigParam) (bool, error)
	searchConfigFn  func(param vo.SearchConfigParam) (*model.ConfigPage, error)
}

func (m *mockConfigClient) GetConfig(param vo.ConfigParam) (string, error) {
	if m.getConfigFn != nil {
		return m.getConfigFn(param)
	}
	return "", nil
}

func (m *mockConfigClient) PublishConfig(param vo.ConfigParam) (bool, error) {
	if m.publishConfigFn != nil {
		return m.publishConfigFn(param)
	}
	return true, nil
}

func (m *mockConfigClient) DeleteConfig(param vo.ConfigParam) (bool, error) {
	if m.deleteConfigFn != nil {
		return m.deleteConfigFn(param)
	}
	return true, nil
}

func (m *mockConfigClient) SearchConfig(param vo.SearchConfigParam) (*model.ConfigPage, error) {
	if m.searchConfigFn != nil {
		return m.searchConfigFn(param)
	}
	return &model.ConfigPage{}, nil
}

func (m *mockConfigClient) CloseClient() {}

func TestConfigGet_JSON(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{
			getConfigFn: func(param vo.ConfigParam) (string, error) {
				if param.DataId != "d1" || param.Group != "g1" {
					t.Fatalf("unexpected param: %+v", param)
				}
				return "content-1", nil
			},
		}, nil
	}

	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "get", "--data-id", "d1", "--group", "g1", "--output", "json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("json unmarshal: %v, output=%s", err, out.String())
	}
	if got["dataId"] != "d1" || got["group"] != "g1" || got["content"] != "content-1" {
		t.Fatalf("unexpected output: %+v", got)
	}
}

func TestConfigList_InvalidSearch(t *testing.T) {
	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "list", "--search", "bad"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestConfigList_DefaultSearchBlur(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{
			searchConfigFn: func(param vo.SearchConfigParam) (*model.ConfigPage, error) {
				if param.Search != "blur" {
					t.Fatalf("unexpected search: %s", param.Search)
				}
				return &model.ConfigPage{TotalCount: 0, PageNumber: 1, PagesAvailable: 0}, nil
			},
		}, nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestConfigGet_UsesResolvedNamespace(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgDir := filepath.Join(home, ".config", "nacos-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"nacos_namespace":"file-ns"}`), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		if cfg.Namespace != "file-ns" {
			t.Fatalf("unexpected namespace: %s", cfg.Namespace)
		}
		return &mockConfigClient{getConfigFn: func(param vo.ConfigParam) (string, error) {
			return "ok", nil
		}}, nil
	}

	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "get", "--data-id", "d1", "--group", "g1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.String() != "ok\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestConfigGet_FlagNamespaceOverridesFile(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgDir := filepath.Join(home, ".config", "nacos-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"nacos_namespace":"file-ns"}`), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		if cfg.Namespace != "flag-ns" {
			t.Fatalf("unexpected namespace: %s", cfg.Namespace)
		}
		return &mockConfigClient{getConfigFn: func(param vo.ConfigParam) (string, error) {
			return "ok", nil
		}}, nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--namespace", "flag-ns", "config", "get", "--data-id", "d1", "--group", "g1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestConfigList_TextTable(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{
			searchConfigFn: func(param vo.SearchConfigParam) (*model.ConfigPage, error) {
				return &model.ConfigPage{
					TotalCount:     2,
					PageNumber:     1,
					PagesAvailable: 1,
					PageItems: []model.ConfigItem{
						{DataId: "d1", Group: "g1"},
						{DataId: "d2", Group: "g2"},
					},
				}, nil
			},
		}, nil
	}

	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Total: 2  Page: 1/1") {
		t.Fatalf("missing summary: %q", got)
	}
	if !strings.Contains(got, "DATA_ID") || !strings.Contains(got, "GROUP") {
		t.Fatalf("missing headers: %q", got)
	}
	if !strings.Contains(got, "d1") || !strings.Contains(got, "g1") {
		t.Fatalf("missing row data: %q", got)
	}
}

func TestConfigGet_EmptyContentNoExtraBlankLine(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{getConfigFn: func(param vo.ConfigParam) (string, error) {
			return "", nil
		}}, nil
	}

	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "get", "--data-id", "d1", "--group", "g1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.String() != "" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestConfigGet_KnownFallbackErrorNormalized(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	home := t.TempDir()
	t.Setenv("HOME", home)

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{getConfigFn: func(param vo.ConfigParam) (string, error) {
			return "", errors.New("read config from both server and cache fail, err=timeout,dataId=d1, group=g1, namespaceId=public")
		}}, nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "get", "--data-id", "d1", "--group", "g1"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	got := err.Error()
	if got != "get config failed, dataId=d1, group=g1, namespace=public, reason=server unavailable and local cache missing" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestConfigGet_PositionalArgs(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	home := t.TempDir()
	t.Setenv("HOME", home)

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{getConfigFn: func(param vo.ConfigParam) (string, error) {
			if param.DataId != "dt-rpc" || param.Group != "COMMON" {
				t.Fatalf("unexpected param: %+v", param)
			}
			return "ok", nil
		}}, nil
	}

	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "get", "dt-rpc", "COMMON"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.String() != "ok\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestConfigGet_DefaultGroupCommon(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	home := t.TempDir()
	t.Setenv("HOME", home)

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{getConfigFn: func(param vo.ConfigParam) (string, error) {
			if param.DataId != "dt-rpc" || param.Group != "COMMON" {
				t.Fatalf("unexpected param: %+v", param)
			}
			return "ok", nil
		}}, nil
	}

	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "get", "dt-rpc"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.String() != "ok\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestCompleteConfigGetArgs(t *testing.T) {
	old := newConfigClient
	defer func() { newConfigClient = old }()

	home := t.TempDir()
	t.Setenv("HOME", home)

	newConfigClient = func(cfg internalconfig.Runtime) (client internalclient.ConfigClient, err error) {
		return &mockConfigClient{searchConfigFn: func(param vo.SearchConfigParam) (*model.ConfigPage, error) {
			return &model.ConfigPage{PageItems: []model.ConfigItem{
				{DataId: "dt-rpc", Group: "COMMON"},
				{DataId: "dt-infra", Group: "COMMON"},
				{DataId: "biz-kv", Group: "COMMON"},
			}}, nil
		}}, nil
	}

	root := NewRootCommand()
	getCmd, _, err := root.Find([]string{"config", "get"})
	if err != nil {
		t.Fatalf("find get command: %v", err)
	}

	dataIDs, directive := completeConfigGetArgs(getCmd, nil, "dt-")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", directive)
	}
	if len(dataIDs) != 2 {
		t.Fatalf("unexpected dataId completion: %#v", dataIDs)
	}

	groups, directive := completeConfigGetArgs(getCmd, []string{"dt-rpc"}, "CO")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", directive)
	}
	if len(groups) != 1 || groups[0] != "COMMON" {
		t.Fatalf("unexpected group completion: %#v", groups)
	}
}
