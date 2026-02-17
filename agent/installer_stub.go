//go:build !installer

package main

import "github.com/spf13/cobra"

func registerInstallerCommands(_ *cobra.Command) {}
