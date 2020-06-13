package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "wayback",
		Short: "A CLI tool for wayback webpages.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(telegramCmd)
}
