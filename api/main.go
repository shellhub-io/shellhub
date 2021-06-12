package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "api"}

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(workerCmd)

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
