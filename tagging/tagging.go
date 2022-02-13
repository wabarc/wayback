// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package tagging // import "github.com/wabarc/wayback/tagging"

import (
	"context"
	"strings"

	"github.com/go-shiori/go-readability"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/zoomio/tagify"
	"github.com/zoomio/tagify/model"
)

var (
	pre = `#`
	sep = ` `
)

// Annotation represents tags associated with webpage content.
type Annotation []string

// Retrieve returns tags associated with the content of a webpage. HTML content
// is preferred, but text content is used if it cannot be found.
func Retrieve(ctx context.Context, art readability.Article) (Annotation, error) {
	parse := func(result *model.Result, err error) (Annotation, bool) {
		if err != nil {
			return Annotation{}, false
		}
		return result.TagsStrings(), true
	}
	opts := []tagify.Option{
		tagify.NoStopWords(true),
		tagify.Limit(config.Opts.MaxTagSize()),
	}

	fromText := func() (Annotation, bool) {
		opts = append(opts, tagify.TargetType(tagify.Text), tagify.Content(art.TextContent))
		return parse(tagify.Run(ctx, opts...))
	}
	fromHTML := func() (Annotation, bool) {
		opts = append(opts, tagify.TargetType(tagify.HTML), tagify.Content(art.Content))
		return parse(tagify.Run(ctx, opts...))
	}

	if an, ok := fromText(); ok && len(an) > 0 {
		return an, nil
	}
	if an, ok := fromHTML(); ok && len(an) > 0 {
		return an, nil
	}

	return Annotation{}, errors.New("retrieve tag failed")
}

// String parses tags and returns a series of tags with the # prefix.
func (ann Annotation) String() string {
	for i, elem := range ann {
		elem = strings.TrimSpace(elem)
		ann[i] = pre + strings.TrimPrefix(elem, pre)
	}
	tags := strings.Join(ann, sep)

	if tags != "" {
		return sep + tags
	}
	return tags
}
