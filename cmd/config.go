package cmd

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"nacos-cli/internal/client"
	internalconfig "nacos-cli/internal/config"
	"nacos-cli/internal/output"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var newConfigClient = client.NewConfigClient

type configListCacheEntry struct {
	items  []model.ConfigItem
	expire time.Time
}

var configListCache struct {
	mu sync.Mutex
	m  map[string]configListCacheEntry
}

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "config operations",
	}

	cmd.AddCommand(newConfigGetCommand())
	cmd.AddCommand(newConfigPutCommand())
	cmd.AddCommand(newConfigDeleteCommand())
	cmd.AddCommand(newConfigListCommand())
	return cmd
}

func newConfigGetCommand() *cobra.Command {
	var dataID string
	var group string

	cmd := &cobra.Command{
		Use:               "get [data-id] [group]",
		Short:             "get config",
		ValidArgsFunction: completeConfigGetArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedDataID := strings.TrimSpace(dataID)
			resolvedGroup := strings.TrimSpace(group)
			if resolvedDataID == "" && len(args) > 0 {
				resolvedDataID = strings.TrimSpace(args[0])
			}
			if resolvedGroup == "" {
				if len(args) > 1 {
					resolvedGroup = strings.TrimSpace(args[1])
				} else {
					resolvedGroup = "COMMON"
				}
			}

			if resolvedDataID == "" {
				return fmt.Errorf("data-id is required")
			}
			if resolvedGroup == "" {
				return fmt.Errorf("group is required")
			}

			runtime, err := internalconfig.ResolveFromCommand(cmd)
			if err != nil {
				return err
			}

			cli, err := newConfigClient(runtime)
			if err != nil {
				return err
			}
			defer cli.CloseClient()

			content, err := cli.GetConfig(vo.ConfigParam{DataId: resolvedDataID, Group: resolvedGroup})
			if err != nil {
				return output.NormalizeConfigGetError(err, resolvedDataID, resolvedGroup, runtime.Namespace)
			}

			return output.Render(cmd.OutOrStdout(), runtime.Output, content, map[string]any{
				"dataId":  resolvedDataID,
				"group":   resolvedGroup,
				"content": content,
			})
		},
	}

	cmd.Flags().StringVar(&dataID, "data-id", "", "config data id")
	cmd.Flags().StringVar(&group, "group", "", "config group")
	return cmd
}

func newConfigPutCommand() *cobra.Command {
	var dataID string
	var group string
	var content string

	cmd := &cobra.Command{
		Use:   "put",
		Short: "put config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(dataID) == "" {
				return fmt.Errorf("data-id is required")
			}
			if strings.TrimSpace(group) == "" {
				return fmt.Errorf("group is required")
			}
			if content == "" {
				return fmt.Errorf("content is required")
			}

			runtime, err := internalconfig.ResolveFromCommand(cmd)
			if err != nil {
				return err
			}

			cli, err := newConfigClient(runtime)
			if err != nil {
				return err
			}
			defer cli.CloseClient()

			ok, err := cli.PublishConfig(vo.ConfigParam{DataId: dataID, Group: group, Content: content})
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("publish config failed")
			}

			return output.Render(cmd.OutOrStdout(), runtime.Output, "ok", map[string]any{"success": true})
		},
	}

	cmd.Flags().StringVar(&dataID, "data-id", "", "config data id")
	cmd.Flags().StringVar(&group, "group", "", "config group")
	cmd.Flags().StringVar(&content, "content", "", "config content")
	_ = cmd.MarkFlagRequired("data-id")
	_ = cmd.MarkFlagRequired("group")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

func newConfigDeleteCommand() *cobra.Command {
	var dataID string
	var group string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(dataID) == "" {
				return fmt.Errorf("data-id is required")
			}
			if strings.TrimSpace(group) == "" {
				return fmt.Errorf("group is required")
			}

			runtime, err := internalconfig.ResolveFromCommand(cmd)
			if err != nil {
				return err
			}

			cli, err := newConfigClient(runtime)
			if err != nil {
				return err
			}
			defer cli.CloseClient()

			ok, err := cli.DeleteConfig(vo.ConfigParam{DataId: dataID, Group: group})
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("delete config failed")
			}

			return output.Render(cmd.OutOrStdout(), runtime.Output, "ok", map[string]any{"success": true})
		},
	}

	cmd.Flags().StringVar(&dataID, "data-id", "", "config data id")
	cmd.Flags().StringVar(&group, "group", "", "config group")
	_ = cmd.MarkFlagRequired("data-id")
	_ = cmd.MarkFlagRequired("group")
	return cmd
}

func completeConfigGetArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	runtimeCmd := cmd
	if cmd.Root() != nil {
		runtimeCmd = cmd.Root()
	}
	runtime, err := internalconfig.ResolveFromCommand(runtimeCmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	items, err := getConfigListItems(runtime)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) == 0 {
		return uniqueDataIDs(items, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	return groupsByDataID(items, args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
}

func getConfigListItems(runtime internalconfig.Runtime) ([]model.ConfigItem, error) {
	cacheKey := runtime.ServerAddr + "|" + runtime.Namespace + "|" + runtime.Username
	now := time.Now()
	configListCache.mu.Lock()
	if configListCache.m == nil {
		configListCache.m = map[string]configListCacheEntry{}
	}
	if entry, ok := configListCache.m[cacheKey]; ok && now.Before(entry.expire) {
		items := make([]model.ConfigItem, len(entry.items))
		copy(items, entry.items)
		configListCache.mu.Unlock()
		return items, nil
	}
	configListCache.mu.Unlock()

	cli, err := newConfigClient(runtime)
	if err != nil {
		return nil, err
	}
	defer cli.CloseClient()

	result, err := cli.SearchConfig(vo.SearchConfigParam{Search: "blur", PageNo: 1, PageSize: 1000})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	items := make([]model.ConfigItem, 0, len(result.PageItems))
	items = append(items, result.PageItems...)

	configListCache.mu.Lock()
	configListCache.m[cacheKey] = configListCacheEntry{items: items, expire: time.Now().Add(10 * time.Second)}
	configListCache.mu.Unlock()
	return items, nil
}

func uniqueDataIDs(items []model.ConfigItem, prefix string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.DataId)
		if id == "" || !strings.HasPrefix(id, prefix) {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func groupsByDataID(items []model.ConfigItem, dataID string, prefix string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, item := range items {
		if item.DataId != dataID {
			continue
		}
		group := strings.TrimSpace(item.Group)
		if group == "" || !strings.HasPrefix(group, prefix) {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		result = append(result, group)
	}
	if len(result) == 0 && strings.HasPrefix("COMMON", prefix) {
		return []string{"COMMON"}
	}
	return result
}

func newConfigListCommand() *cobra.Command {
	var search string
	var dataID string
	var group string
	var pageNo int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if search != "accurate" && search != "blur" {
				return fmt.Errorf("search must be one of: accurate, blur")
			}

			runtime, err := internalconfig.ResolveFromCommand(cmd)
			if err != nil {
				return err
			}

			cli, err := newConfigClient(runtime)
			if err != nil {
				return err
			}
			defer cli.CloseClient()

			result, err := cli.SearchConfig(vo.SearchConfigParam{
				Search:   search,
				DataId:   dataID,
				Group:    group,
				PageNo:   pageNo,
				PageSize: pageSize,
			})
			if err != nil {
				return err
			}
			if result == nil {
				if runtime.Output == "json" {
					return output.Render(cmd.OutOrStdout(), runtime.Output, "", map[string]any{
						"totalCount":     0,
						"pageNumber":     pageNo,
						"pagesAvailable": 0,
						"pageItems":      []any{},
					})
				}
				return output.Render(cmd.OutOrStdout(), runtime.Output, output.RenderTable(
					fmt.Sprintf("Total: 0  Page: %d/0", pageNo),
					[]string{"DATA_ID", "GROUP"},
					nil,
				), nil)
			}

			if runtime.Output == "json" {
				return output.Render(cmd.OutOrStdout(), runtime.Output, "", result)
			}

			rows := make([][]string, 0, len(result.PageItems))
			for _, item := range result.PageItems {
				rows = append(rows, []string{item.DataId, item.Group})
			}
			return output.Render(cmd.OutOrStdout(), runtime.Output, output.RenderTable(
				fmt.Sprintf("Total: %d  Page: %d/%d", result.TotalCount, result.PageNumber, result.PagesAvailable),
				[]string{"DATA_ID", "GROUP"},
				rows,
			), nil)
		},
	}

	cmd.Flags().StringVar(&search, "search", "blur", "search mode: accurate|blur")
	cmd.Flags().StringVar(&dataID, "data-id", "", "config data id")
	cmd.Flags().StringVar(&group, "group", "", "config group")
	cmd.Flags().IntVar(&pageNo, "page-no", 1, "page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 10, "page size")
	return cmd
}
