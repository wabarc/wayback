// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.
package main

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/spf13/cobra"
	"github.com/wabarc/helper"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/reduxer"
	"golang.org/x/sync/errgroup"
)

func assets(art reduxer.Artifact) []reduxer.Asset {
	return []reduxer.Asset{
		art.Img,
		art.PDF,
		art.Raw,
		art.Txt,
		art.HAR,
		art.WARC,
		art.Media,
	}
}

func archive(cmd *cobra.Command, args []string) {
	// TODO: clean the auto-created temporary directory.
	archiving := func(ctx context.Context, urls []*url.URL) error {
		g, ctx := errgroup.WithContext(ctx)
		rdx, err := reduxer.Do(ctx, urls...)
		if err != nil {
			return errors.Wrap(err, "reduxer unexpected")
		}
		cols, err := wayback.Wayback(ctx, rdx, urls...)
		if err != nil {
			return err
		}

		content := pretty(cols, rdx)
		for _, line := range strings.Split(content, "\n") {
			cmd.Println(line)
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

func pretty(cols []wayback.Collect, rdx reduxer.Reduxer) string {
	writer := list.NewWriter()
	defer writer.Reset()

	type uri string
	type collects []wayback.Collect
	grouped := make(map[uri]collects, len(cols)/4)
	for _, col := range cols {
		src := uri(col.Src)
		grouped[src] = append(grouped[src], col)
	}

	for src := range grouped {
		writer.AppendItem(src)
		writer.Indent()
		items := make([]interface{}, 0)
		for _, col := range grouped[src] {
			item := fmt.Sprintf("%s: %s", strings.ToUpper(col.Arc), col.Dst)
			items = append(items, item)
		}

		hasArtifact := false
		artifacts := make([]interface{}, 0)
		if bundle, ok := rdx.Load(reduxer.Src(src)); ok {
			for _, asset := range assets(bundle.Artifact()) {
				hasArtifact = true
				if asset.Local == "" {
					continue
				}
				artifacts = append(artifacts, asset.Local)
			}
		}
		if hasArtifact {
			items = append(items, "Artifacts")
		}
		writer.AppendItems(items)
		if hasArtifact {
			writer.Indent()
			writer.AppendItems(artifacts)
			writer.UnIndent()
		}
		writer.UnIndent()
	}
	writer.SetStyle(list.StyleConnectedRounded)

	return writer.Render()
}

func unmarshalArgs(args []string) (urls []*url.URL, err error) {
	for _, s := range args {
		if helper.IsURL(s) {
			uri, er := url.Parse(s)
			if er != nil {
				err = errors.Wrap(er, "parse url failed")
				continue
			}
			urls = append(urls, uri)
		} else {
			uris := readFromFile(s)
			if len(uris) > 0 {
				urls = append(urls, uris...)
			}
		}
	}
	return
}

func readFromFile(s string) (urls []*url.URL) {
	if helper.Exists(s) {
		file, err := os.Open(s)
		if err != nil {
			return
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			uri, err := url.Parse(scanner.Text())
			if err == nil {
				urls = append(urls, uri)
			}
		}
	}
	return
}
