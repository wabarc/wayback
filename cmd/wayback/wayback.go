package main

import (
	"github.com/spf13/cobra"
	"github.com/wabarc/wayback"
)

func wbia(cmd *cobra.Command, args []string) r {
	wbrc = &wayback.Handle{URI: args}

	return wbrc.IA()
}

func wbis(cmd *cobra.Command, args []string) r {
	wbrc = &wayback.Handle{URI: args}

	return wbrc.IS()
}

func wbip(cmd *cobra.Command, args []string) r {
	wbrc = &wayback.Handle{
		URI:  args,
		IPFS: new(wayback.IPFSRV),
	}

	return wbrc.WBIPFS()
}
