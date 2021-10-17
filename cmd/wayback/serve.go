package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/pooling"
	"github.com/wabarc/wayback/service/discord"
	"github.com/wabarc/wayback/service/httpd"
	"github.com/wabarc/wayback/service/mastodon"
	"github.com/wabarc/wayback/service/matrix"
	"github.com/wabarc/wayback/service/relaychat"
	"github.com/wabarc/wayback/service/slack"
	"github.com/wabarc/wayback/service/telegram"
	"github.com/wabarc/wayback/service/twitter"
	"github.com/wabarc/wayback/storage"
	"github.com/wabarc/wayback/systemd"
)

type target struct {
	call func()
	name string
}

type service struct {
	targets []target
}

func serve(_ *cobra.Command, _ []string) {
	store, err := storage.Open("")
	if err != nil {
		logger.Fatal("open storage failed: %v", err)
	}
	defer store.Close()

	pool := pooling.New(config.Opts.PoolingSize())
	defer pool.Close()

	ctx, cancel := context.WithCancel(context.Background())
	srv := &service{}
	_ = srv.run(ctx, store, pool)

	if systemd.HasNotifySocket() {
		logger.Info("sending readiness notification to Systemd")

		if err := systemd.SdNotify(systemd.SdNotifyReady); err != nil {
			logger.Error("unable to send readiness notification to systemd: %v", err)
		}
	}

	go srv.stop(cancel)
	<-ctx.Done()

	logger.Info("wayback service stopped.")
}

// nolint:gocyclo
func (srv *service) run(ctx context.Context, store *storage.Storage, pool pooling.Pool) *service {
	size := len(daemon)
	srv.targets = make([]target, 0, size)
	for _, s := range daemon {
		switch s {
		case "irc":
			irc := relaychat.New(ctx, store, pool)
			go func() {
				if err := irc.Serve(); err != relaychat.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { irc.Shutdown() },
				name: s,
			})
		case "slack":
			sl := slack.New(ctx, store, pool)
			go func() {
				if err := sl.Serve(); err != slack.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { sl.Shutdown() },
				name: s,
			})
		case "discord":
			d := discord.New(ctx, store, pool)
			go func() {
				if err := d.Serve(); err != discord.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { d.Shutdown() },
				name: s,
			})
		case "mastodon", "mstdn":
			m := mastodon.New(ctx, store, pool)
			go func() {
				if err := m.Serve(); err != mastodon.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { m.Shutdown() },
				name: s,
			})
		case "telegram":
			t := telegram.New(ctx, store, pool)
			go func() {
				if err := t.Serve(); err != telegram.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { t.Shutdown() },
				name: s,
			})
		case "twitter":
			t := twitter.New(ctx, store, pool)
			go func() {
				if err := t.Serve(); err != twitter.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { t.Shutdown() },
				name: s,
			})
		case "matrix":
			m := matrix.New(ctx, store, pool)
			go func() {
				if err := m.Serve(); err != matrix.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { m.Shutdown() },
				name: s,
			})
		case "web", "httpd":
			h := httpd.New(ctx, store, pool)
			go func() {
				if err := h.Serve(); err != httpd.ErrServiceClosed {
					logger.Error("%v", err)
				}
			}()
			srv.targets = append(srv.targets, target{
				call: func() { h.Shutdown() },
				name: s,
			})
		default:
			logger.Error("unrecognize %s in `--daemon`", s)
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

	var once sync.Once
	for {
		sig := <-signalChan
		if sig == os.Interrupt {
			logger.Info("Signal SIGINT is received, probably due to `Ctrl-C`, exiting...")
			once.Do(func() {
				srv.shutdown()
				cancel()
			})
			return
		}
	}
}

func (srv *service) shutdown() {
	for _, target := range srv.targets {
		logger.Info("stopping %s service...", target.name)
		target.call()
	}
}
