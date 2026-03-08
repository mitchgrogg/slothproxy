package main

import (
	"fmt"

	"github.com/mitchgrogg/rita-devtools/slothproxy/pkg/slothproxy"
	"github.com/spf13/cobra"
)

func buildVersionCommand(_ *slothproxy.SlothProxy) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), slothproxy.Version)
			return nil
		},
	}
}
