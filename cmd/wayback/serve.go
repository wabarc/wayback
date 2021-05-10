package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/service/anonymity"
	"github.com/wabarc/wayback/service/mastodon"
	"github.com/wabarc/wayback/service/matrix"
	"github.com/wabarc/wayback/service/relaychat"
	"github.com/wabarc/wayback/service/telegram"
	"github.com/wabarc/wayback/service/twitter"
	"github.com/wabarc/wayback/storage"
)

type service struct {
	errCh chan error
}

func serve(_ *cobra.Command, args []string) {
	store, err := storage.Open("")
	if err != nil {
		logger.Fatal("open storage failed: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	srv := &service{}
	ran := srv.run(ctx, store)

	go srv.stop(cancel)
	defer close(srv.errCh)

	select {
	case err := <-ran.err():
		logger.Error("%v", err.Error())
	case <-ctx.Done():
		time.Sleep(100 * time.Millisecond)
		logger.Info("wayback service stopped.")
	}
}

func (srv *service) run(ctx context.Context, store *storage.Storage) *service {
	srv.errCh = make(chan error, len(daemon))
	for _, s := range daemon {
		switch s {
		case "irc":
			irc := relaychat.New(store)
			go func(errCh chan error) {
				errCh <- irc.Serve(ctx)
			}(srv.errCh)
		case "mastodon", "mstdn":
			mastodon := mastodon.New(store)
			go func(errCh chan error) {
				errCh <- mastodon.Serve(ctx)
			}(srv.errCh)
		case "telegram":
			telegram := telegram.New(store)
			go func(errCh chan error) {
				errCh <- telegram.Serve(ctx)
			}(srv.errCh)
		case "twitter":
			twitter := twitter.New(store)
			go func(errCh chan error) {
				errCh <- twitter.Serve(ctx)
			}(srv.errCh)
		case "matrix":
			matrix := matrix.New(store)
			go func(errCh chan error) {
				errCh <- matrix.Serve(ctx)
			}(srv.errCh)
		case "web":
			tor := anonymity.New(store)
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
			logger.Info("Signal SIGINT is received, probably due to `Ctrl-C`, exiting...")
			cancel()
			return
		}
	}
}

func (s *service) err() <-chan error { return s.errCh }
