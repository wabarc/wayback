package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wabarc/wayback"
)

type r = map[string]string

type m struct {
	ia r
	is r
	ip r
}

var (
	ia bool
	is bool
	ip bool

	daemon string

	host string
	port uint
	mode string
	tor  bool

	token  string
	chatid string
	debug  bool

	rootCmd = &cobra.Command{
		Use:   "wayback",
		Short: "A CLI tool for wayback webpages.",
		Example: `  wayback https://www.wikipedia.org
  wayback https://www.fsf.org https://www.eff.org
  wayback --ia https://www.fsf.org
  wayback --ip https://www.fsf.org
  wayback --ia --is -d telegram -t your-telegram-bot-token
  WAYBACK_SLOT=pinata WAYBACK_APIKEY=YOUR-PINATA-APIKEY \
    WAYBACK_SECRET=YOUR-PINATA-SECRET wayback --ip https://www.fsf.org`,
		Run: func(cmd *cobra.Command, args []string) {
			// Assign default flags
			assign()
			handle(cmd, args)
		},
		Version: "1.0.0",
	}
)

var wbrc wayback.Broker

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&ia, "ia", "", false, "Wayback webpages to Internet Archive.")
	rootCmd.Flags().BoolVarP(&is, "is", "", false, "Wayback webpages to Archive Today.")
	rootCmd.Flags().BoolVarP(&ip, "ip", "", false, "Wayback webpages to IPFS. (default false)")
	rootCmd.Flags().StringVarP(&daemon, "daemon", "d", "", "Run as daemon service.")
	rootCmd.Flags().StringVarP(&host, "ipfs-host", "", "127.0.0.1", "IPFS daemon host, do not require, unless enable ipfs.")
	rootCmd.Flags().UintVarP(&port, "ipfs-port", "p", 5001, "IPFS daemon port.")
	rootCmd.Flags().StringVarP(&mode, "ipfs-mode", "m", "pinner", "IPFS mode.")
	rootCmd.Flags().BoolVarP(&tor, "tor", "", false, "Snapshot webpage use tor proxy.")

	rootCmd.Flags().StringVarP(&token, "token", "t", "", "Telegram bot API Token, required.")
	rootCmd.Flags().StringVarP(&chatid, "chatid", "c", "", "Channel ID. default: \"\"")
	rootCmd.Flags().BoolVarP(&debug, "debug", "", false, "Enable debug mode. default: false")
}

func assign() {
	if !ia && !is && !ip {
		ia, is = true, true
	}
}

func output(tit string, args r) {
	fmt.Printf("[%s]\n", tit)
	for ori, dst := range args {
		fmt.Printf("%s => %s", ori, dst)
	}
	fmt.Print("\n")
}

func handle(cmd *cobra.Command, args []string) {
	switch daemon {
	case "telegram":
		telegram()
	default:
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}

		r := &m{}
		if ia {
			r.ia = wbia(cmd, args)
			output("Internet Archive", r.ia)
		}
		if is {
			r.is = wbis(cmd, args)
			output("Archive Today", r.is)
		}
		if ip {
			r.ip = wbip(cmd, args)
			output("IPFS", r.ip)
		}
	}
}

func telegram() {
	ipfs := &wayback.IPFSRV{Host: host, Port: port, Mode: mode, UseTor: tor}
	handle := map[string]bool{
		"ia": ia,
		"is": is,
		"ip": ip,
	}
	wbrc := wayback.NewConfig(token, debug, chatid, handle, ipfs)

	wbrc.Telegram()
}
