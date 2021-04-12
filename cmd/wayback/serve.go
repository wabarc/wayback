package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wayback/service/anonymity"
	"github.com/wabarc/wayback/service/mastodon"
	"github.com/wabarc/wayback/service/matrix"
	"github.com/wabarc/wayback/service/relaychat"
	"github.com/wabarc/wayback/service/telegram"
	"github.com/wabarc/wayback/service/twitter"
)

type service struct {
	errCh chan error
}

func serve(_ *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := &service{}
	ran := srv.run(ctx)

	go srv.stop(cancel)
	defer close(srv.errCh)

	select {
	case err := <-ran.err():
		logger.Error("%v", err.Error())
	case <-ctx.Done():
		logger.Info("Wayback service stopped.")
	}
}

func (srv *service) run(ctx context.Context) *service {
	srv.errCh = make(chan error, len(daemon))
	for _, s := range daemon {
		switch s {
		case "irc":
			irc := relaychat.New()
			go func(errCh chan error) {
				errCh <- irc.Serve(ctx)
			}(srv.errCh)
		case "mastodon", "mstdn":
			mastodon := mastodon.New()
			go func(errCh chan error) {
				errCh <- mastodon.Serve(ctx)
			}(srv.errCh)
		case "telegram":
			telegram := telegram.New()
			go func(errCh chan error) {
				errCh <- telegram.Serve(ctx)
			}(srv.errCh)
		case "twitter":
			twitter := twitter.New()
			go func(errCh chan error) {
				errCh <- twitter.Serve(ctx)
			}(srv.errCh)
		case "matrix":
			matrix := matrix.New()
			go func(errCh chan error) {
				errCh <- matrix.Serve(ctx)
			}(srv.errCh)
		case "web":
			tor := anonymity.New()
			go func(errCh chan error) {
				errCh <- tor.Serve(ctx)
			}(srv.errCh)
		default:
			fmt.Printf("Unrecognize %s in `--daemon`\n", s)
			srv.errCh <- ctx.Err()
		}
	}

	return srv
}

func (srv *service) stop(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		os.Interrupt,
	)

	for {
		sig := <-signalChan
		switch sig {
		case os.Interrupt:
			logger.Info("Signal SIGINT is received, probably due to `Ctrl-C`, exiting ...")
			cancel()
			return
		}
	}
}

func (s *service) err() <-chan error { return s.errCh }
