package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/version"
)

var (
	err error

	ia bool
	is bool
	ip bool
	ph bool

	daemon []string

	host string
	port uint
	mode string
	tor  bool

	token  string
	chatid string
	torKey string

	debug bool
	info  bool
	print bool

	configFile string

	rootCmd = &cobra.Command{
		Use:   "wayback",
		Short: "A CLI tool for wayback webpages.",
		Example: `  wayback https://www.wikipedia.org
  wayback https://www.fsf.org https://www.eff.org
  wayback --ia https://www.fsf.org
  wayback --ia --is -d telegram -t your-telegram-bot-token
  WAYBACK_SLOT=pinata WAYBACK_APIKEY=YOUR-PINATA-APIKEY \
    WAYBACK_SECRET=YOUR-PINATA-SECRET wayback --ip https://www.fsf.org`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkRequiredFlags(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			handle(cmd, args)
		},
		Version: version.Version,
	}
)

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&ia, "ia", "", false, "Wayback webpages to Internet Archive.")
	rootCmd.Flags().BoolVarP(&is, "is", "", false, "Wayback webpages to Archive Today.")
	rootCmd.Flags().BoolVarP(&ip, "ip", "", false, "Wayback webpages to IPFS. (default false)")
	rootCmd.Flags().BoolVarP(&ph, "ph", "", false, "Wayback webpages to Telegraph. (default false)")
	rootCmd.Flags().StringSliceVarP(&daemon, "daemon", "d", []string{}, "Run as daemon service, supported services are telegram, web, mastodon, twitter, discord, slack, irc")
	rootCmd.Flags().StringVarP(&host, "ipfs-host", "", "127.0.0.1", "IPFS daemon host, do not require, unless enable ipfs.")
	rootCmd.Flags().UintVarP(&port, "ipfs-port", "p", 5001, "IPFS daemon port.")
	rootCmd.Flags().StringVarP(&mode, "ipfs-mode", "m", "pinner", "IPFS mode.")
	rootCmd.Flags().BoolVarP(&tor, "tor", "", false, "Snapshot webpage via Tor anonymity network.")

	rootCmd.Flags().StringVarP(&token, "token", "t", "", "Telegram Bot API Token.")
	rootCmd.Flags().StringVarP(&chatid, "chatid", "", "", "Telegram channel id.")
	rootCmd.Flags().StringVarP(&torKey, "tor-key", "", "", "The private key for Tor Hidden Service.")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path, defaults: ./wayback.conf, ~/wayback.conf, /etc/wayback.conf")
	rootCmd.Flags().BoolVarP(&debug, "debug", "", false, "Enable debug mode. (default false)")
	rootCmd.Flags().BoolVarP(&info, "info", "", false, "Show application information.")
	rootCmd.Flags().BoolVarP(&print, "print", "", false, "Show application configurations.")
}

func checkRequiredFlags(cmd *cobra.Command) error {
	flags := cmd.Flags()
	for _, d := range daemon {
		switch d {
		case "telegram":
			if flags.Changed("token") && strings.TrimSpace(token) == "" {
				return errors.New("Token of the Telegram Bot is required to run as Telegram service.")
			}
		case "web":
			if flags.Changed("tor-key") && strings.TrimSpace(torKey) == "" {
				return errors.New("The private key for Tor service is required.")
			}
		}
	}

	if flags.Changed("chatid") && strings.TrimSpace(chatid) == "" {
		return errors.New("Telegram Channel name is required with flag `--chatid` or `-c`.")
	}

	return nil
}

func setToEnv(cmd *cobra.Command) {
	flags := cmd.Flags()

	if flags.Changed("debug") {
		os.Setenv("DEBUG", fmt.Sprint(debug))
	}
	if flags.Changed("ia") {
		os.Setenv("WAYBACK_ENABLE_IA", fmt.Sprint(ia))
	}
	if flags.Changed("is") {
		os.Setenv("WAYBACK_ENABLE_IS", fmt.Sprint(is))
	}
	if flags.Changed("ip") {
		os.Setenv("WAYBACK_ENABLE_IP", fmt.Sprint(ip))
	}
	if flags.Changed("ph") {
		os.Setenv("WAYBACK_ENABLE_PH", fmt.Sprint(ph))
	}
	if flags.Changed("token") {
		os.Setenv("WAYBACK_TELEGRAM_TOKEN", token)
	}
	if flags.Changed("chatid") {
		os.Setenv("WAYBACK_TELEGRAM_CHANNEL", chatid)
	}
	if flags.Changed("host") {
		os.Setenv("WAYBACK_IPFS_HOST", host)
	}
	if flags.Changed("port") {
		os.Setenv("WAYBACK_IPFS_PORT", fmt.Sprint(port))
	}
	if flags.Changed("mode") {
		os.Setenv("WAYBACK_IPFS_MODE", mode)
	}
	if flags.Changed("tor") {
		os.Setenv("WAYBACK_USE_TOR", fmt.Sprint(tor))
	}
	if flags.Changed("tor-key") {
		os.Setenv("WAYBACK_TOR_PRIVKEY", torKey)
	}
}

// nolint:gocyclo
func handle(cmd *cobra.Command, args []string) {
	if !ia && !is && !ip && !ph {
		ia, is = true, true
		os.Setenv("WAYBACK_ENABLE_IA", "true")
		os.Setenv("WAYBACK_ENABLE_IS", "true")
	}

	setToEnv(cmd)
	parser := config.NewParser()

	if configFile != "" {
		if len(daemon) > 0 {
			logger.Info("Run wayback using configuration file")
		}
		if config.Opts, err = parser.ParseFile(configFile); err != nil {
			logger.Fatal("Parse configuration file failed, error: %v", err)
		}
	}

	if config.Opts, err = parser.ParseEnvironmentVariables(); err != nil {
		logger.Fatal("Parse environment variables or flags failed, error: %v", err)
	}

	if !config.Opts.LogTime() {
		logger.DisableTime()
	}

	logger.SetLogLevel(config.Opts.LogLevel())
	if debug || config.Opts.HasDebugMode() {
		logger.EnableDebug()
	}

	if config.Opts.EnabledMetrics() {
		metrics.Gather = metrics.NewCollector()
	}

	if info {
		showInfo(cmd)
		return
	}

	if print {
		cmd.Println(spew.Sdump(config.Opts))
		return
	}

	hasDaemon := len(daemon) > 0
	hasArgs := len(args) > 0
	switch {
	case hasDaemon:
		serve(cmd, args)
	case hasArgs:
		archive(cmd, args)
	default:
		cmd.Help()
	}
	os.Exit(0)
}

func showInfo(cmd *cobra.Command) {
	cmd.Println("Version:", version.Version)
	cmd.Println("Commit:", version.Commit)
	cmd.Println("Build Date:", version.BuildDate)
	cmd.Println("Go Version:", runtime.Version())
	cmd.Println("Compiler:", runtime.Compiler)
	cmd.Println("Arch:", runtime.GOARCH)
	cmd.Println("OS:", runtime.GOOS)
}
