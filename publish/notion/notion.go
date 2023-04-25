// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package notion // import "github.com/wabarc/wayback/publish/notion"

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
	"github.com/dstotijn/go-notion"
	"github.com/wabarc/helper"
	"github.com/wabarc/imgbb"
	"github.com/wabarc/logger"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/publish"
	"github.com/wabarc/wayback/reduxer"
	"github.com/wabarc/wayback/template/render"
	"golang.org/x/net/html"
)

// Interface guard
var _ publish.Publisher = (*Notion)(nil)

type Notion struct {
	bot  *notion.Client
	opts *config.Options
}

// New returns a notion client.
func New(client *http.Client, opts *config.Options) *Notion {
	if opts.NotionToken() == "" {
		logger.Debug("Notion integration access token is required")
		return nil
	}

	bot := notion.NewClient(opts.NotionToken())
	if client != nil {
		newcli := notion.WithHTTPClient(client)
		bot = notion.NewClient(opts.NotionToken(), newcli)
	}

	return &Notion{bot: bot, opts: opts}
}

// Publish publish text to the Notion block of the given cols and args.
// A context should contain a `reduxer.Reduxer` via `publish.PubBundle` struct.
//
// TODO rate limit https://developers.notion.com/reference/request-limits
func (no *Notion) Publish(ctx context.Context, rdx reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishNotion, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishNotion, metrics.StatusFailure)
		return errors.New("publish to notion: collects empty")
	}

	var head = render.Title(cols, rdx)
	var body = render.ForPublish(&render.Notion{Cols: cols, Data: rdx}).String()
	if head == "" {
		head = "Published at " + time.Now().Format("2006-01-02T15:04:05")
	}

	params := no.params(cols, head, body)
	if err := params.Validate(); err != nil {
		return errors.Wrap(err, "notion page params invalid")
	}

	page, err := no.bot.CreatePage(ctx, params)
	if err != nil {
		metrics.IncrementPublish(metrics.PublishNotion, metrics.StatusFailure)
		return errors.Wrap(err, "create page failed")
	}

	logger.Debug("created page: %v", page)
	metrics.IncrementPublish(metrics.PublishNotion, metrics.StatusSuccess)
	return nil
}

func (no *Notion) params(cols []wayback.Collect, head, body string) notion.CreatePageParams {
	// tips := "Toggle Archiving"
	table := []notion.Block{}
	for i, col := range cols {
		// Add the source URI to the first row
		if i == 0 {
			row := notion.TableRowBlock{
				Cells: [][]notion.RichText{
					{
						{Text: &notion.Text{Content: "Source"}, Annotations: &notion.Annotations{Bold: true}},
					},
					{
						{Text: &notion.Text{Content: col.Src, Link: &notion.Link{URL: col.Src}}},
					},
				},
			}
			table = append(table, row)
		}
		row := notion.TableRowBlock{
			Cells: [][]notion.RichText{
				{
					{Text: &notion.Text{Content: config.SlotName(col.Arc)}, Annotations: &notion.Annotations{Bold: true}},
				},
				{
					{Text: &notion.Text{Content: col.Dst, Link: &notion.Link{URL: col.Dst}}},
				},
			},
		}
		table = append(table, row)
	}

	children := []notion.Block{
		notion.DividerBlock{},
		notion.TableBlock{
			TableWidth:   2,
			HasRowHeader: true,
			Children:     table,
		},
		// notion.ToggleBlock{
		// 	RichText: []notion.RichText{
		// 		{
		// 			Text: &notion.Text{
		// 				Content: tips,
		// 			},
		// 		},
		// 	},
		// 	Children: []notion.Block{
		// 		notion.TableBlock{
		// 			TableWidth:   2,
		// 			HasRowHeader: true,
		// 			Children:     table,
		// 		},
		// 	},
		// },
		notion.DividerBlock{},
		// {
		// 	Object: "block",
		// 	Type:   notion.BlockTypeHeading2,
		// 	Heading2: &notion.Heading{
		// 		Text: []notion.RichText{
		// 			notion.RichText{
		// 				Type: notion.RichTextTypeText,
		// 				Text: &notion.Text{
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

	params := notion.CreatePageParams{
		ParentType: notion.ParentTypeDatabase,
		ParentID:   no.opts.NotionDatabaseID(),
		DatabasePageProperties: &notion.DatabasePageProperties{
			"title": notion.DatabasePageProperty{
				Title: []notion.RichText{
					{
						Type: notion.RichTextTypeText,
						Text: &notion.Text{
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

func traverseNodes(selections *goquery.Selection, client *imgbb.ImgBB) []notion.Block {
	var element notion.Block
	var blocks []notion.Block
	selections.Each(func(_ int, child *goquery.Selection) {
		for _, node := range child.Nodes {
			switch node.Type {
			case html.TextNode:
				if len(strings.TrimSpace(node.Data)) > 0 {
					element = notion.ParagraphBlock{
						RichText: []notion.RichText{
							{
								Text: &notion.Text{
									Content: html.EscapeString(node.Data),
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
							element = notion.ImageBlock{
								Type: notion.FileTypeExternal,
								External: &notion.FileExternal{
									URL: attr.Val,
								},
							}
							blocks = append(blocks, element)
						}
					}
					// case "pre":
					// 	element = notion.Block{
					// 		Object: "block",
					// 		Type:   notion.BlockTypeCode,
					// 		Code: &notion.Code{
					// 			RichTextBlock: notion.RichTextBlock{
					//                 Text: []notion.RichText{},
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
