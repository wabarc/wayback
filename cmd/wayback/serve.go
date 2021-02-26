package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/service/anonymity"
	"github.com/wabarc/wayback/service/mastodon"
	"github.com/wabarc/wayback/service/telegram"
)

type service struct {
	errCh chan error
}

func serve(cmd *cobra.Command, opts *config.Options, args []string) {
	ctx := context.Background()
	srv := &service{}
	ran := srv.run(ctx, opts)

	select {
	case err := <-ran.err():
		cmd.Println(err.Error())
	case <-ctx.Done():
	}
}

func (srv *service) run(ctx context.Context, opts *config.Options) *service {
	srv.errCh = make(chan error, len(daemon))
	for _, s := range daemon {
		switch s {
		case "telegram":
			telegram := telegram.New(opts)
			go func(errCh chan error) {
				errCh <- telegram.Serve(ctx)
			}(srv.errCh)
		case "mastodon", "mstdn":
			mastodon := mastodon.New(opts)
			go func(errCh chan error) {
				errCh <- mastodon.Serve(ctx)
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

func (s *service) err() <-chan error { return s.errCh }
