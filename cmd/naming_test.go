package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/cobra"
	internalclient "nacos-cli/internal/client"
	internalconfig "nacos-cli/internal/config"
)

type mockNamingClient struct {
	registerFn   func(param vo.RegisterInstanceParam) (bool, error)
	deregisterFn func(param vo.DeregisterInstanceParam) (bool, error)
	instancesFn  func(param vo.SelectInstancesParam) ([]model.Instance, error)
}

func (m *mockNamingClient) RegisterInstance(param vo.RegisterInstanceParam) (bool, error) {
	if m.registerFn != nil {
		return m.registerFn(param)
	}
	return true, nil
}

func (m *mockNamingClient) DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error) {
	if m.deregisterFn != nil {
		return m.deregisterFn(param)
	}
	return true, nil
}

func (m *mockNamingClient) SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error) {
	if m.instancesFn != nil {
		return m.instancesFn(param)
	}
	return nil, nil
}

func (m *mockNamingClient) CloseClient() {}

func TestNamingInstances_Text(t *testing.T) {
	old := newNamingClient
	defer func() { newNamingClient = old }()

	newNamingClient = func(cfg internalconfig.Runtime) (client internalclient.NamingClient, err error) {
		return &mockNamingClient{
			instancesFn: func(param vo.SelectInstancesParam) ([]model.Instance, error) {
				if param.ServiceName != "svc1" {
					t.Fatalf("unexpected service: %s", param.ServiceName)
				}
				return []model.Instance{{Ip: "1.1.1.1", Port: 8080}}, nil
			},
		}, nil
	}

	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"naming", "instances", "--service", "svc1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := out.String()
	if got == "" || !bytes.Contains([]byte(got), []byte("count: 1")) || !bytes.Contains([]byte(got), []byte("1.1.1.1:8080")) {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestNamingRegister_Validate(t *testing.T) {
	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"naming", "register", "--service", "", "--ip", "1.1.1.1", "--port", "8080"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCompleteNamespace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".config", "nacos-cli"), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "nacos-cli", "config.json"), []byte(`{"namespaces":["prepare_hwc","online_hwc","job_hwc"]}`), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	got, directive := completeNamespace(nil, nil, "pre")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", directive)
	}
	if len(got) != 1 || got[0] != "prepare_hwc" {
		t.Fatalf("unexpected completion: %#v", got)
	}
}
