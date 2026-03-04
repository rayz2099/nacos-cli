package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveFromCommand_Defaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := testCommand()

	r, err := ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}

	if r.ServerAddr != DefaultServerAddr {
		t.Fatalf("ServerAddr = %s", r.ServerAddr)
	}
	if r.Namespace != DefaultNamespace {
		t.Fatalf("Namespace = %s", r.Namespace)
	}
	if r.Output != DefaultOutput {
		t.Fatalf("Output = %s", r.Output)
	}
}

func TestResolveFromCommand_Env(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("nacos_server_addr", "10.0.0.1:8848")
	t.Setenv("nacos_username", "u1")
	t.Setenv("nacos_password", "p1")
	t.Setenv("nacos_namespace", "ns1")
	t.Setenv("nacos_output", "json")

	cmd := testCommand()
	r, err := ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}

	if r.ServerAddr != "10.0.0.1:8848" || r.Username != "u1" || r.Password != "p1" || r.Namespace != "ns1" || r.Output != "json" {
		t.Fatalf("resolved runtime mismatch: %+v", r)
	}
}

func TestResolveFromCommand_FlagOverridesEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("nacos_server_addr", "10.0.0.1:8848")
	t.Setenv("nacos_output", "text")
	t.Setenv("nacos_namespace", "env-ns")

	cmd := testCommand()
	if err := cmd.Flags().Set("server-addr", "127.0.0.1:9999"); err != nil {
		t.Fatalf("set server-addr: %v", err)
	}
	if err := cmd.Flags().Set("output", "json"); err != nil {
		t.Fatalf("set output: %v", err)
	}
	if err := cmd.Flags().Set("namespace", "flag-ns"); err != nil {
		t.Fatalf("set namespace: %v", err)
	}

	r, err := ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}

	if r.ServerAddr != "127.0.0.1:9999" {
		t.Fatalf("ServerAddr = %s", r.ServerAddr)
	}
	if r.Output != "json" {
		t.Fatalf("Output = %s", r.Output)
	}
	if r.Namespace != "flag-ns" {
		t.Fatalf("Namespace = %s", r.Namespace)
	}
}

func TestResolveFromCommand_InvalidOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cmd := testCommand()
	if err := cmd.Flags().Set("output", "yaml"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	_, err := ResolveFromCommand(cmd)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestResolveFromCommand_FileConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "nacos-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	content := []byte(`{"nacos_server_addr":"192.168.1.10:8848","nacos_username":"fu","nacos_password":"fp","nacos_namespace":"fns","namespaces":["prepare_hwc","online_hwc"],"nacos_output":"json"}`)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cmd := testCommand()
	r, err := ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}

	if r.ServerAddr != "192.168.1.10:8848" || r.Username != "fu" || r.Password != "fp" || r.Namespace != "fns" || r.Output != "json" {
		t.Fatalf("resolved runtime mismatch: %+v", r)
	}
}

func TestResolveFromCommand_FlagOverridesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "nacos-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	content := []byte(`{"nacos_server_addr":"192.168.1.10:8848","nacos_namespace":"file-ns","nacos_output":"text"}`)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cmd := testCommand()
	if err := cmd.Flags().Set("server-addr", "127.0.0.1:9999"); err != nil {
		t.Fatalf("set server-addr: %v", err)
	}
	if err := cmd.Flags().Set("output", "json"); err != nil {
		t.Fatalf("set output: %v", err)
	}
	if err := cmd.Flags().Set("namespace", "flag-ns"); err != nil {
		t.Fatalf("set namespace: %v", err)
	}

	r, err := ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}

	if r.ServerAddr != "127.0.0.1:9999" {
		t.Fatalf("ServerAddr = %s", r.ServerAddr)
	}
	if r.Output != "json" {
		t.Fatalf("Output = %s", r.Output)
	}
	if r.Namespace != "flag-ns" {
		t.Fatalf("Namespace = %s", r.Namespace)
	}
}

func TestResolveFromCommand_Priority_FlagOverEnvOverFileOverDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "nacos-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	content := []byte(`{"nacos_namespace":"file-ns"}`)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cmd := testCommand()
	r, err := ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}
	if r.Namespace != "file-ns" {
		t.Fatalf("Namespace from file = %s", r.Namespace)
	}

	t.Setenv("nacos_namespace", "env-ns")
	r, err = ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}
	if r.Namespace != "env-ns" {
		t.Fatalf("Namespace from env = %s", r.Namespace)
	}

	if err := cmd.Flags().Set("namespace", "flag-ns"); err != nil {
		t.Fatalf("set namespace: %v", err)
	}
	r, err = ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}
	if r.Namespace != "flag-ns" {
		t.Fatalf("Namespace from flag = %s", r.Namespace)
	}
}

func TestResolveFromCommand_NamespacesFirstAsDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "nacos-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	content := []byte(`{"namespaces":["prepare_hwc","online_hwc","public"]}`)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cmd := testCommand()
	r, err := ResolveFromCommand(cmd)
	if err != nil {
		t.Fatalf("ResolveFromCommand() error = %v", err)
	}
	if r.Namespace != "prepare_hwc" {
		t.Fatalf("Namespace = %s", r.Namespace)
	}
}

func TestNamespaceCandidates_FromFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "nacos-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	content := []byte(`{"nacos_namespace":"runtime","namespaces":["prepare_hwc","online_hwc","public"]}`)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	got := NamespaceCandidates()
	if len(got) < 4 {
		t.Fatalf("unexpected candidates: %#v", got)
	}
	if got[0] != "prepare_hwc" {
		t.Fatalf("unexpected first candidate: %#v", got)
	}
}

func testCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("server-addr", "", "")
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("namespace", "", "")
	cmd.Flags().String("output", "", "")
	return cmd
}
