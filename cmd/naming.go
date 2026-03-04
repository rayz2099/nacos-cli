package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"nacos-cli/internal/client"
	internalconfig "nacos-cli/internal/config"
	"nacos-cli/internal/output"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var newNamingClient = client.NewNamingClient

func newNamingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "naming",
		Short: "naming operations",
	}

	cmd.AddCommand(newNamingRegisterCommand())
	cmd.AddCommand(newNamingDeregisterCommand())
	cmd.AddCommand(newNamingInstancesCommand())
	return cmd
}

func newNamingRegisterCommand() *cobra.Command {
	var service string
	var ip string
	var port uint64
	var group string
	var cluster string
	var weight float64
	var ephemeral bool

	cmd := &cobra.Command{
		Use:   "register",
		Short: "register instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(service) == "" {
				return fmt.Errorf("service is required")
			}
			if strings.TrimSpace(ip) == "" {
				return fmt.Errorf("ip is required")
			}
			if port == 0 {
				return fmt.Errorf("port is required")
			}

			runtime, err := internalconfig.ResolveFromCommand(cmd)
			if err != nil {
				return err
			}

			cli, err := newNamingClient(runtime)
			if err != nil {
				return err
			}
			defer cli.CloseClient()

			ok, err := cli.RegisterInstance(vo.RegisterInstanceParam{
				Ip:          ip,
				Port:        port,
				Weight:      weight,
				Enable:      true,
				Healthy:     true,
				ClusterName: cluster,
				ServiceName: service,
				GroupName:   group,
				Ephemeral:   ephemeral,
			})
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("register instance failed")
			}

			return output.Render(cmd.OutOrStdout(), runtime.Output, "ok", map[string]any{"success": true})
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "service name")
	cmd.Flags().StringVar(&ip, "ip", "", "instance ip")
	cmd.Flags().Uint64Var(&port, "port", 0, "instance port")
	cmd.Flags().StringVar(&group, "group", constant.DEFAULT_GROUP, "group name")
	cmd.Flags().StringVar(&cluster, "cluster", "DEFAULT", "cluster name")
	cmd.Flags().Float64Var(&weight, "weight", 1, "instance weight")
	cmd.Flags().BoolVar(&ephemeral, "ephemeral", true, "ephemeral instance")
	_ = cmd.MarkFlagRequired("service")
	_ = cmd.MarkFlagRequired("ip")
	_ = cmd.MarkFlagRequired("port")
	return cmd
}

func newNamingDeregisterCommand() *cobra.Command {
	var service string
	var ip string
	var port uint64
	var group string
	var cluster string
	var ephemeral bool

	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "deregister instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(service) == "" {
				return fmt.Errorf("service is required")
			}
			if strings.TrimSpace(ip) == "" {
				return fmt.Errorf("ip is required")
			}
			if port == 0 {
				return fmt.Errorf("port is required")
			}

			runtime, err := internalconfig.ResolveFromCommand(cmd)
			if err != nil {
				return err
			}

			cli, err := newNamingClient(runtime)
			if err != nil {
				return err
			}
			defer cli.CloseClient()

			ok, err := cli.DeregisterInstance(vo.DeregisterInstanceParam{
				Ip:          ip,
				Port:        port,
				Cluster:     cluster,
				ServiceName: service,
				GroupName:   group,
				Ephemeral:   ephemeral,
			})
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("deregister instance failed")
			}

			return output.Render(cmd.OutOrStdout(), runtime.Output, "ok", map[string]any{"success": true})
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "service name")
	cmd.Flags().StringVar(&ip, "ip", "", "instance ip")
	cmd.Flags().Uint64Var(&port, "port", 0, "instance port")
	cmd.Flags().StringVar(&group, "group", constant.DEFAULT_GROUP, "group name")
	cmd.Flags().StringVar(&cluster, "cluster", "DEFAULT", "cluster name")
	cmd.Flags().BoolVar(&ephemeral, "ephemeral", true, "ephemeral instance")
	_ = cmd.MarkFlagRequired("service")
	_ = cmd.MarkFlagRequired("ip")
	_ = cmd.MarkFlagRequired("port")
	return cmd
}

func newNamingInstancesCommand() *cobra.Command {
	var service string
	var group string
	var clusters string
	var healthyOnly bool

	cmd := &cobra.Command{
		Use:   "instances",
		Short: "list service instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(service) == "" {
				return fmt.Errorf("service is required")
			}

			runtime, err := internalconfig.ResolveFromCommand(cmd)
			if err != nil {
				return err
			}

			cli, err := newNamingClient(runtime)
			if err != nil {
				return err
			}
			defer cli.CloseClient()

			result, err := cli.SelectInstances(vo.SelectInstancesParam{
				ServiceName: service,
				GroupName:   group,
				Clusters:    splitClusters(clusters),
				HealthyOnly: healthyOnly,
			})
			if err != nil {
				return err
			}

			if runtime.Output == "json" {
				return output.Render(cmd.OutOrStdout(), runtime.Output, "", result)
			}

			lines := make([]string, 0, len(result)+1)
			lines = append(lines, fmt.Sprintf("count: %d", len(result)))
			for _, item := range result {
				lines = append(lines, item.Ip+":"+strconv.FormatUint(item.Port, 10))
			}
			return output.Render(cmd.OutOrStdout(), runtime.Output, strings.Join(lines, "\n"), nil)
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "service name")
	cmd.Flags().StringVar(&group, "group", constant.DEFAULT_GROUP, "group name")
	cmd.Flags().StringVar(&clusters, "clusters", "", "comma separated clusters")
	cmd.Flags().BoolVar(&healthyOnly, "healthy-only", true, "only healthy instances")
	_ = cmd.MarkFlagRequired("service")
	return cmd
}

func splitClusters(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
