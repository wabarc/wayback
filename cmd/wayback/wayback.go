package main

import (
	"context"
	"fmt"
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
	archiving := func(ctx context.Context, urls []string) error {
		g, ctx := errgroup.WithContext(ctx)
		cols, err := wayback.Wayback(ctx, &bundles, urls...)
		if err != nil {
			return err
		}

		for _, col := range cols {
			cmd.Println(col.Src, "=>", col.Dst)
		}
		for src, bundle := range bundles {
			for _, path := range bundle.Paths() {
				if path == "" {
					continue
				}
				cmd.Println(src, "=>", path)
			}
		}

		if err := g.Wait(); err != nil {
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := archiving(ctx, args); err != nil {
		cmd.PrintErrln(err)
	}
}
