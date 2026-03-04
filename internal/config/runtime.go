package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	DefaultServerAddr = "127.0.0.1:8848"
	DefaultNamespace  = "public"
	DefaultOutput     = "text"
	DefaultConfigPath = ".config/nacos-cli/config.json"
)

type Runtime struct {
	ServerAddr string
	Username   string
	Password   string
	Namespace  string
	Output     string
}

type fileConfig struct {
	ServerAddr string   `json:"nacos_server_addr"`
	Username   string   `json:"nacos_username"`
	Password   string   `json:"nacos_password"`
	Namespace  string   `json:"nacos_namespace"`
	Namespaces []string `json:"namespaces"`
	Output     string   `json:"nacos_output"`
}

func ResolveFromCommand(cmd *cobra.Command) (Runtime, error) {
	fileCfg, err := loadFileConfig()
	if err != nil {
		return Runtime{}, err
	}

	serverAddr, err := resolveString(cmd, "server-addr", []string{"nacos_server_addr", "NACOS_SERVER_ADDR"}, fileCfg.ServerAddr, DefaultServerAddr)
	if err != nil {
		return Runtime{}, err
	}

	username, err := resolveString(cmd, "username", []string{"nacos_username", "NACOS_USERNAME"}, fileCfg.Username, "")
	if err != nil {
		return Runtime{}, err
	}

	password, err := resolveString(cmd, "password", []string{"nacos_password", "NACOS_PASSWORD"}, fileCfg.Password, "")
	if err != nil {
		return Runtime{}, err
	}

	namespace, err := resolveNamespace(cmd, fileCfg)
	if err != nil {
		return Runtime{}, err
	}

	output, err := resolveString(cmd, "output", []string{"nacos_output", "NACOS_OUTPUT"}, fileCfg.Output, DefaultOutput)
	if err != nil {
		return Runtime{}, err
	}
	output = strings.ToLower(output)
	if output != "text" && output != "json" {
		return Runtime{}, fmt.Errorf("invalid output: %s", output)
	}

	if strings.TrimSpace(serverAddr) == "" {
		return Runtime{}, fmt.Errorf("server-addr is required")
	}

	return Runtime{
		ServerAddr: serverAddr,
		Username:   username,
		Password:   password,
		Namespace:  namespace,
		Output:     output,
	}, nil
}

func resolveString(cmd *cobra.Command, flagName string, envKeys []string, fileValue string, defaultValue string) (string, error) {
	value, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return "", err
	}
	flag := cmd.Flags().Lookup(flagName)
	if flag != nil && flag.Changed {
		return value, nil
	}

	for _, key := range envKeys {
		envValue, ok := os.LookupEnv(key)
		if ok {
			return envValue, nil
		}
	}

	if strings.TrimSpace(fileValue) != "" {
		return fileValue, nil
	}

	return defaultValue, nil
}

func resolveNamespace(cmd *cobra.Command, fileCfg fileConfig) (string, error) {
	value, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return "", err
	}
	flag := cmd.Flags().Lookup("namespace")
	if flag != nil && flag.Changed {
		return value, nil
	}

	for _, key := range []string{"nacos_namespace", "NACOS_NAMESPACE"} {
		envValue, ok := os.LookupEnv(key)
		if ok {
			return envValue, nil
		}
	}

	if strings.TrimSpace(fileCfg.Namespace) != "" {
		return fileCfg.Namespace, nil
	}
	for _, item := range fileCfg.Namespaces {
		v := strings.TrimSpace(item)
		if v != "" {
			return v, nil
		}
	}

	return DefaultNamespace, nil
}

func NamespaceCandidates() []string {
	cfg, err := loadFileConfig()
	if err != nil {
		return []string{DefaultNamespace}
	}

	result := make([]string, 0, len(cfg.Namespaces)+2)
	seen := map[string]struct{}{}
	appendUnique := func(value string) {
		v := strings.TrimSpace(value)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}

	for _, item := range cfg.Namespaces {
		appendUnique(item)
	}
	appendUnique(cfg.Namespace)
	appendUnique(DefaultNamespace)
	if len(result) == 0 {
		return []string{DefaultNamespace}
	}
	return result
}

func loadFileConfig() (fileConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fileConfig{}, nil
	}

	path := filepath.Join(homeDir, DefaultConfigPath)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileConfig{}, nil
		}
		return fileConfig{}, err
	}

	cfg := fileConfig{}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return fileConfig{}, fmt.Errorf("invalid config file %s: %w", path, err)
	}

	return cfg, nil
}
