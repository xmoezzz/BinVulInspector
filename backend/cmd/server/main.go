package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"bin-vul-inspector/cmd/version"
)

var (
	rootCmd = &cobra.Command{
		Version:      fmt.Sprintf("%s, commit %s, date %s", version.Version, version.GitCommit, version.BuildTime),
		Use:          "bin-vul-inspector",
		Short:        "bin-vul-inspector",
		Long:         "Bin-Vul-Inspector Server and Tools",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
)

func init() {
	rootCmd.AddCommand(New())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
