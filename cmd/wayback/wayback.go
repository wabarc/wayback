package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/reduxer"
	"golang.org/x/sync/errgroup"
)

func output(tit string, args map[string]string) {
	fmt.Printf("[%s]\n", tit)
	for ori, dst := range args {
		fmt.Printf("%s => %s\n", ori, dst)
	}
}

func archive(cmd *cobra.Command, args []string) {
	var bundles reduxer.Bundles
	archiving := func(ctx context.Context, urls []*url.URL) error {
		g, ctx := errgroup.WithContext(ctx)
		cols, err := wayback.Wayback(ctx, &bundles, urls...)
		if err != nil {
			return err
		}

		for _, col := range cols {
			cmd.Println(col.Src, "=>", col.Dst)
		}
		for src, bundle := range bundles {
			for _, asset := range bundle.Asset() {
				if asset.Local == "" {
					continue
				}
				cmd.Println(src, "=>", asset.Local)
			}
		}

		if err := g.Wait(); err != nil {
			return err
		}
		return nil
	}

	urls, err := unmarshalArgs(args)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := archiving(ctx, urls); err != nil {
		cmd.PrintErrln(err)
	}
}

func unmarshalArgs(args []string) (urls []*url.URL, err error) {
	for _, s := range args {
		uri, er := url.Parse(s)
		if er != nil {
			err = fmt.Errorf("%w: unexpect url: %s", err, s)
			continue
		}
		urls = append(urls, uri)
	}
	return
}
