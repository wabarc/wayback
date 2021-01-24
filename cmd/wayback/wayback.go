package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"golang.org/x/sync/errgroup"
)

func output(tit string, args map[string]string) {
	fmt.Printf("[%s]\n", tit)
	for ori, dst := range args {
		fmt.Printf("%s => %s", ori, dst)
	}
	fmt.Print("\n")
}

func archive(cmd *cobra.Command, opts *config.Options, args []string) {
	archiving := func(ctx context.Context, urls []string) error {
		g, ctx := errgroup.WithContext(ctx)
		var wbrc wayback.Broker = &wayback.Handle{URLs: urls, Opts: opts}

		for slot, do := range opts.Slots() {
			slot, do := slot, do
			g.Go(func() error {
				switch {
				case slot == config.SLOT_IA && do:
					output(config.SlotName(config.SLOT_IA), wbrc.IA())
				case slot == config.SLOT_IS && do:
					output(config.SlotName(config.SLOT_IS), wbrc.IS())
				case slot == config.SLOT_IP && do:
					output(config.SlotName(config.SLOT_IP), wbrc.IP())
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := archiving(ctx, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
