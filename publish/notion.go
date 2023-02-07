// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package publish // import "github.com/wabarc/wayback/publish"

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/wabarc/helper"
	"github.com/wabarc/imgbb"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/template/render"
	"golang.org/x/net/html"

	notionapi "github.com/dstotijn/go-notion"
)

var _ Publisher = (*notion)(nil)

type notion struct {
	client *notionapi.Client
}

// NewNotion returns a notion client.
func NewNotion(httpClient *http.Client) *notion {
	if config.Opts.NotionToken() == "" {
		logger.Error("Notion integration access token is required")
		return new(notion)
	}

	client := notionapi.NewClient(config.Opts.NotionToken())
	if httpClient != nil {
		opts := notionapi.WithHTTPClient(httpClient)
		client = notionapi.NewClient(config.Opts.NotionToken(), opts)
	}

	return &notion{client: client}
}

// Publish publish text to the Notion block of the given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
//
// TODO rate limit https://developers.notion.com/reference/request-limits
func (no *notion) Publish(ctx context.Context, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishNotion, metrics.StatusRequest)
	if no.client == nil {
		return errors.New("publish to notion failed: client nil")
	}

	if len(cols) == 0 {
		return errors.New("publish to notion: collects empty")
	}

	rdx, _, err := extract(ctx, cols)
	if err != nil {
		logger.Warn("extract data failed: %v", err)
	}

	var head = render.Title(cols, rdx)
	var body = render.ForPublish(&render.Notion{Cols: cols, Data: rdx}).String()
	if head == "" {
		head = "Published at " + time.Now().Format("2006-01-02T15:04:05")
	}

	params := no.params(cols, head, body)
	if err = params.Validate(); err != nil {
		return errors.Wrap(err, "notion page params invalid")
	}

	page, err := no.client.CreatePage(ctx, params)
	if err != nil {
		metrics.IncrementPublish(metrics.PublishNotion, metrics.StatusFailure)
		return errors.Wrap(err, "create page failed")
	}

	logger.Debug("created page: %v", page)
	metrics.IncrementPublish(metrics.PublishNotion, metrics.StatusSuccess)
	return nil
}

func (no *notion) params(cols []wayback.Collect, head, body string) notionapi.CreatePageParams {
	// tips := "Toggle open archived targets."
	table := []notionapi.Block{}
	for i, col := range cols {
		// Add the source URI to the first row
		if i == 0 {
			row := notionapi.Block{
				TableRow: &notionapi.TableRow{
					Cells: [][]notionapi.RichText{
						{
							{Type: notionapi.RichTextTypeText, Text: &notionapi.Text{Content: "Source"}, Annotations: &notionapi.Annotations{Bold: true}},
						},
						{
							{Type: notionapi.RichTextTypeText, Text: &notionapi.Text{Content: col.Src, Link: &notionapi.Link{URL: col.Src}}},
						},
					},
				},
			}
			table = append(table, row)
		}
		row := notionapi.Block{
			TableRow: &notionapi.TableRow{
				Cells: [][]notionapi.RichText{
					{
						{Type: notionapi.RichTextTypeText, Text: &notionapi.Text{Content: config.SlotName(col.Arc)}, Annotations: &notionapi.Annotations{Bold: true}},
					},
					{
						{Type: notionapi.RichTextTypeText, Text: &notionapi.Text{Content: col.Dst, Link: &notionapi.Link{URL: col.Dst}}},
					},
				},
			},
		}
		table = append(table, row)
	}

	children := []notionapi.Block{
		{
			Object:  "object",
			Type:    notionapi.BlockTypeDivider,
			Divider: &notionapi.Divider{},
		},
		{
			Object: "object",
			Type:   notionapi.BlockTypeTable, // TODO replace with toggle list
			Table: &notionapi.Table{
				TableWidth:   2,
				HasRowHeader: true,
				Children:     table,
			},
		},
		{
			Object:  "object",
			Type:    notionapi.BlockTypeDivider,
			Divider: &notionapi.Divider{},
		},
		// {
		// 	Object: "block",
		// 	Type:   notionapi.BlockTypeHeading2,
		// 	Heading2: &notionapi.Heading{
		// 		Text: []notionapi.RichText{
		// 			notionapi.RichText{
		// 				Type: notionapi.RichTextTypeText,
		// 				Text: &notionapi.Text{
		// 					Content: head,
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}

	if doc, err := goquery.NewDocumentFromReader(strings.NewReader(body)); err == nil {
		nodes := traverseNodes(doc.Contents(), imgbb.NewImgBB(nil, ""))
		children = append(children, nodes...)
	}

	params := notionapi.CreatePageParams{
		ParentType: notionapi.ParentTypeDatabase,
		ParentID:   config.Opts.NotionDatabaseID(),
		DatabasePageProperties: &notionapi.DatabasePageProperties{
			"Name": notionapi.DatabasePageProperty{
				Title: []notionapi.RichText{
					{
						Type: notionapi.RichTextTypeText,
						Text: &notionapi.Text{
							Content: head,
						},
					},
				},
			},
		},
		Children: children,
	}

	return params
}

func traverseNodes(selections *goquery.Selection, client *imgbb.ImgBB) []notionapi.Block {
	var element notionapi.Block
	var blocks []notionapi.Block
	selections.Each(func(_ int, child *goquery.Selection) {
		for _, node := range child.Nodes {
			switch node.Type {
			case html.TextNode:
				if len(strings.TrimSpace(node.Data)) > 0 {
					element = notionapi.Block{
						Object: "block",
						Type:   notionapi.BlockTypeParagraph,
						Paragraph: &notionapi.RichTextBlock{
							Text: []notionapi.RichText{
								{
									Type: notionapi.RichTextTypeText,
									Text: &notionapi.Text{
										Content: html.EscapeString(node.Data),
									},
								},
							},
						},
					}
					blocks = append(blocks, element)
				}
			case html.ElementNode:
				switch node.Data {
				case "img":
					for _, attr := range node.Attr {
						if attr.Key == "src" && strings.TrimSpace(attr.Val) != "" {
							// Upload the image to a third-party image hosting service
							newurl, err := uploadImage(client, attr.Val)
							if err == nil {
								attr.Val = newurl
							}
							element = notionapi.Block{
								Object: "block",
								Type:   notionapi.BlockTypeImage,
								Image: &notionapi.FileBlock{
									Type: notionapi.FileTypeExternal,
									External: &notionapi.FileExternal{
										URL: attr.Val,
									},
								},
							}
							blocks = append(blocks, element)
						}
					}
					// case "pre":
					// 	element = notionapi.Block{
					// 		Object: "block",
					// 		Type:   notionapi.BlockTypeCode,
					// 		Code: &notionapi.Code{
					// 			RichTextBlock: notionapi.RichTextBlock{
					//                 Text: []notionapi.RichText{},
					// 				Children: traverseNodes(child.Contents(), client),
					// 			},
					// 		},
					// 	}
					// 	blocks = append(blocks, element)
				default:
				}
			}
		}
		blocks = append(blocks, traverseNodes(child.Contents(), client)...)
	})

	return blocks
}

func download(u *url.URL) (path string, err error) {
	path = filepath.Join(os.TempDir(), helper.RandString(21, "lower"))
	fd, err := os.Create(path)
	if err != nil {
		return path, err
	}
	defer fd.Close()

	resp, err := http.Get(u.String()) // nosemgrep: gitlab.gosec.G104-1.G107-1, gitlab.gosec.G107-1, gitlab.gosec.G108-1
	if err != nil {
		return path, err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(fd, resp.Body); err != nil {
		return path, err
	}

	return path, nil
}

func uploadImage(client *imgbb.ImgBB, s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", errors.Wrap(err, "parse url "+s)
	}

	path, err := download(u)
	if err != nil {
		return "", errors.Wrap(err, "download url "+s)
	}
	defer os.Remove(path)

	newurl, err := client.Upload(path)
	if err != nil || newurl == "" {
		return newurl, errors.Wrap(err, "upload file "+path)
	}
	newurl += "?orig=" + s

	return newurl, nil
}
