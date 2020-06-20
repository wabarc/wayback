package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wabarc/wayback"
)

var (
	token  string
	chatid string
	debug  bool

	telegramCmd = &cobra.Command{
		Use:   "telegram",
		Short: "A CLI tool for wayback webpages on Telegram bot.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(token) == 0 {
				cmd.Help()
				os.Exit(0)
			}

			wbrc := wayback.NewConfig(token, debug, chatid)

			wbrc.Telegram()
		},
	}
)

func init() {
	telegramCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "Telegram bot API Token, required.")
	telegramCmd.PersistentFlags().StringVarP(&chatid, "chatid", "c", "", "Channel ID. default: \"\"")
	telegramCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode. default: false")
	telegramCmd.MarkFlagRequired("token")
}
