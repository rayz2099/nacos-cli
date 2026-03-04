package main

import (
	"os"

	"nacos-cli/cmd"
	"nacos-cli/internal/output"
)

func main() {
	root := cmd.NewRootCommand()
	if err := root.Execute(); err != nil {
		output.RenderError(os.Stderr, err)
		os.Exit(1)
	}
}
