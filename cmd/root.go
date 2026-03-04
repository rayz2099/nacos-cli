package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	internalconfig "nacos-cli/internal/config"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "nacos-cli",
		Short:         "nacos command line tool",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().String("server-addr", "", "nacos server address")
	cmd.PersistentFlags().String("username", "", "nacos username")
	cmd.PersistentFlags().String("password", "", "nacos password")
	cmd.PersistentFlags().String("namespace", "", "nacos namespace")
	cmd.PersistentFlags().StringP("output", "o", "", "output format: text|json")
	_ = cmd.RegisterFlagCompletionFunc("namespace", completeNamespace)

	cmd.AddCommand(newConfigCommand())
	cmd.AddCommand(newNamingCommand())
	return cmd
}

func completeNamespace(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	namespaceCandidates := internalconfig.NamespaceCandidates()
	result := make([]string, 0, len(namespaceCandidates))
	for _, item := range namespaceCandidates {
		if strings.HasPrefix(item, toComplete) {
			result = append(result, item)
		}
	}
	return result, cobra.ShellCompDirectiveNoFileComp
}
