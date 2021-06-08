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
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/service/httpd"
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

func serve(_ *cobra.Command, _ []string) {
	store, err := storage.Open("")
	if err != nil {
		logger.Fatal("open storage failed: %v", err)
	}
	defer store.Close()

	pool := pooling.New(config.Opts.PoolingSize())

	ctx, cancel := context.WithCancel(context.Background())
	srv := &service{}
	ran := srv.run(ctx, store, pool)

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

func (srv *service) run(ctx context.Context, store *storage.Storage, pool pooling.Pool) *service {
	srv.errCh = make(chan error, len(daemon))
	for _, s := range daemon {
		switch s {
		case "irc":
			irc := relaychat.New(ctx, store, pool)
			go func(errCh chan error) {
				errCh <- irc.Serve()
			}(srv.errCh)
		case "mastodon", "mstdn":
			mastodon := mastodon.New(ctx, store, pool)
			go func(errCh chan error) {
				errCh <- mastodon.Serve()
			}(srv.errCh)
		case "telegram":
			telegram := telegram.New(ctx, store, pool)
			go func(errCh chan error) {
				errCh <- telegram.Serve()
			}(srv.errCh)
		case "twitter":
			twitter := twitter.New(ctx, store, pool)
			go func(errCh chan error) {
				errCh <- twitter.Serve()
			}(srv.errCh)
		case "matrix":
			matrix := matrix.New(ctx, store, pool)
			go func(errCh chan error) {
				errCh <- matrix.Serve()
			}(srv.errCh)
		case "web":
			tor := httpd.New(ctx, store, pool)
			go func(errCh chan error) {
				errCh <- tor.Serve()
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
		os.Interrupt,
	)

	for {
		sig := <-signalChan
		if sig == os.Interrupt {
			logger.Info("Signal SIGINT is received, probably due to `Ctrl-C`, exiting...")
			cancel()
			return
		}
	}
}

func (srv *service) err() <-chan error { return srv.errCh }
