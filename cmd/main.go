package main

import (
	"github.com/spf13/cobra"
)

var (
	host string
	port uint
	tor  bool

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
	rootCmd.PersistentFlags().StringVarP(&host, "host", "", "127.0.0.1", "IPFS daemon host.")
	rootCmd.PersistentFlags().UintVarP(&port, "port", "", 5001, "IPFS daemon port.")
	rootCmd.PersistentFlags().BoolVarP(&tor, "tor", "", false, "Saving webpage use tor proxy.")
	rootCmd.AddCommand(telegramCmd)
}
