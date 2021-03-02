package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/logger"
	"github.com/wabarc/wayback/service/anonymity"
	"github.com/wabarc/wayback/service/mastodon"
	"github.com/wabarc/wayback/service/telegram"
	"github.com/wabarc/wayback/service/twitter"
)

type service struct {
	errCh chan error
}

func serve(_ *cobra.Command, opts *config.Options, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := &service{}
	ran := srv.run(ctx, opts)

	go srv.stop(cancel)
	defer close(srv.errCh)

	select {
	case err := <-ran.err():
		logger.Error("%v", err.Error())
	case <-ctx.Done():
		logger.Info("Wayback service stopped.")
	}
}

func (srv *service) run(ctx context.Context, opts *config.Options) *service {
	srv.errCh = make(chan error, len(daemon))
	for _, s := range daemon {
		switch s {
		case "mastodon", "mstdn":
			mastodon := mastodon.New(opts)
			go func(errCh chan error) {
				errCh <- mastodon.Serve(ctx)
			}(srv.errCh)
		case "telegram":
			telegram := telegram.New(opts)
			go func(errCh chan error) {
				errCh <- telegram.Serve(ctx)
			}(srv.errCh)
		case "twitter":
			twitter := twitter.New(opts)
			go func(errCh chan error) {
				errCh <- twitter.Serve(ctx)
			}(srv.errCh)
		case "web":
			tor := anonymity.New(opts)
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
