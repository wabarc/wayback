package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wabarc/playback"
	"github.com/wabarc/wayback"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "playback",
		Short:   "A toolkit to playback archived webpage from time capsules.",
		Example: `  playback https://example.com https://example.org`,
		Version: playback.Version,
		Run: func(cmd *cobra.Command, args []string) {
			handle(cmd, args)
		},
	}

	rootCmd.Execute()
}

func handle(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
		os.Exit(0)
	}

	collects, _ := wayback.Playback(context.TODO(), args...)
	for _, collect := range collects {
		fmt.Printf("[%s]\n", collect.Arc)
		for orig, dest := range collect.Dst {
			fmt.Println(orig, "=>", dest)
		}
		fmt.Printf("\n")
	}
}
