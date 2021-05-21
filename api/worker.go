package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use: "worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("worker!")
		return nil
	},
}
