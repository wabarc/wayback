// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package meili // import "github.com/wabarc/wayback/publish/meili"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/rs/xid"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
	"github.com/wabarc/wayback/metrics"
	"github.com/wabarc/wayback/reduxer"
)

var (
	timeout    = 10 * time.Second
	indexing   = `capsules`
	primaryKey = `id`

	userAgent   = `WaybackArchiver/1.0`
	contentType = `application/json`

	errIndexNotFound = errors.New(fmt.Sprintf(`indexing %s not found`, indexing))
	errIndexNotMatch = errors.New(fmt.Sprintf(`indexing %s not match`, indexing))
)

// Meili represents a Meilisearch client.
type Meili struct {
	ctx    context.Context
	client *http.Client

	// Meilisearch server API endpoint.
	endpoint string

	// Meilisearch indexing name.
	indexing string

	// Meilisearch admin API key, which can be emptied.
	apikey string

	// Version of the Meilisearch server.
	version string
}

// New returns a Meilisearch client.
func New(ctx context.Context, client *http.Client, opts *config.Options) *Meili {
	endpoint, apikey, idxname := opts.MeiliEndpoint(), opts.MeiliApikey(), opts.MeiliIndexing()

	if client == nil {
		client = &http.Client{
			Timeout: timeout,
		}
	}
	if idxname == "" {
		idxname = indexing
	}

	return &Meili{
		ctx:      ctx,
		endpoint: endpoint,
		indexing: idxname,
		apikey:   apikey,
		client:   client,
	}
}

// setup creates an index if one does not already exist on the server.
func (m *Meili) setup() error {
	err := m.existIndex()
	if errors.Is(err, errIndexNotFound) {
		err = m.createIndex()
	}
	if err != nil {
		return err
	}
	err = m.getVersion()
	if err != nil {
		return err
	}
	return m.sortable()
}

// getVersion specifies its version of the meilisearch server.
func (m *Meili) getVersion() error {
	endpoint := fmt.Sprintf(`%s/version`, m.endpoint)
	resp, err := m.do(http.MethodGet, endpoint, nil)
	if err != nil {
		return errors.Wrap(err, `get version: request failed`)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(`get version: request failed: ` + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `get version: reads body failed`)
	}

	var server struct {
		Version string `json:"pkgVersion"`
	}
	if err := json.Unmarshal(body, &server); err != nil {
		return errors.Wrap(err, `get version: unmarshal json failed`)
	}
	m.version = server.Version

	return nil
}

type indexes struct {
	UID        string    `json:"uid"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	PrimaryKey string    `json:"primaryKey"`
}

// existIndex returns whether or not the indexing exists.
func (m *Meili) existIndex() error {
	endpoint := fmt.Sprintf(`%s/indexes/%s`, m.endpoint, m.indexing)
	resp, err := m.do(http.MethodGet, endpoint, nil)
	if err != nil {
		return errors.Wrap(err, `get index: request failed`)
	}
	defer resp.Body.Close()

	// Indexing does not exist.
	if resp.StatusCode == http.StatusNotFound {
		return errIndexNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(`get index: request failed: ` + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `get index: reads body failed`)
	}

	var idx indexes
	if err := json.Unmarshal(body, &idx); err != nil {
		return errors.Wrap(err, `get index: unmarshal json failed`)
	}
	if idx.UID != m.indexing {
		return errIndexNotFound
	}

	return nil
}

type creates struct {
	EnqueuedAt time.Time `json:"enqueuedAt"`
	IndexUID   string    `json:"indexUid"`
	Status     string    `json:"status"`
	Type       string    `json:"type"`
	UID        int       `json:"uid"`
}

// createIndex creates an index.
func (m *Meili) createIndex() error {
	endpoint := fmt.Sprintf(`%s/indexes`, m.endpoint)
	payload := fmt.Sprintf(`{"uid":"%s", "primaryKey":"%s"}`, m.indexing, primaryKey)
	resp, err := m.do(http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return errors.Wrap(err, `create index: request failed`)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(`create index: unexpected status: ` + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `create index: reads body failed`)
	}

	var idx creates
	if err := json.Unmarshal(body, &idx); err != nil {
		return errors.Wrap(err, `create index: unmarshal json failed`)
	}
	if idx.IndexUID != m.indexing {
		return errIndexNotMatch
	}

	return nil
}

// sortable sets an index's sortable attributes.
func (m *Meili) sortable() error {
	endpoint := fmt.Sprintf(`%s/indexes/%s/settings/sortable-attributes`, m.endpoint, m.indexing)
	payload := `["id"]`
	method := http.MethodPost
	ver, err := version.NewVersion(m.version)
	if err != nil {
		return errors.Wrap(err, `set sortable attributes: invalid version: `+m.version)
	}
	// The method of updating the searchable attributes settings changed to `PUT`
	// See https://github.com/meili/meili/releases/tag/v0.28.0
	constraints, err := version.NewConstraint(`>= 0.28`)
	if err != nil {
		return errors.Wrap(err, `set sortable attributes: new constraint failed`)
	}
	if constraints.Check(ver) {
		method = http.MethodPut
	}

	resp, err := m.do(method, endpoint, strings.NewReader(payload))
	if err != nil {
		return errors.Wrap(err, `set sortable attributes: request failed`)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(`set sortable attributes: unexpected status: ` + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `set sortable attributes: reads body failed`)
	}

	var idx creates
	if err := json.Unmarshal(body, &idx); err != nil {
		return errors.Wrap(err, `set sortable attributes: unmarshal json failed`)
	}
	if idx.IndexUID != m.indexing {
		return errIndexNotMatch
	}

	return nil
}

type document struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	IA     string `json:"ia"`
	IS     string `json:"is"`
	IP     string `json:"ip"`
	PH     string `json:"ph"`
}

// push documents
func (m *Meili) Publish(_ context.Context, _ reduxer.Reduxer, cols []wayback.Collect, args ...string) error {
	metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusRequest)

	if len(cols) == 0 {
		metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusFailure)
		return errors.New(`push documents failed: cols empty`)
	}

	buf, err := json.Marshal(m.documents(cols))
	if err != nil {
		metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusFailure)
		return errors.Wrap(err, `push document: marshal docs failed`)
	}

	endpoint := fmt.Sprintf(`%s/indexes/%s/documents`, m.endpoint, m.indexing)
	resp, err := m.do(http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusFailure)
		return errors.Wrap(err, `push document: failed`)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusFailure)
		return errors.New(`push document: unexpected status: ` + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusFailure)
		return errors.Wrap(err, `push document: reads body failed`)
	}

	var idx creates
	if err := json.Unmarshal(body, &idx); err != nil {
		metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusFailure)
		return errors.Wrap(err, `push document: unmarshal json failed`)
	}
	if idx.IndexUID != m.indexing {
		metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusFailure)
		return errIndexNotMatch
	}
	metrics.IncrementPublish(metrics.PublishMeili, metrics.StatusSuccess)

	return nil
}

func (m *Meili) do(method, url string, body io.Reader) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, body) // nolint:errcheck
	if m.apikey != "" {
		req.Header.Add("Authorization", "Bearer "+m.apikey)
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("User-Agent", userAgent)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *Meili) documents(cols []wayback.Collect) (docs []document) {
	for src, maps := range groupBySrc(cols) {
		doc := document{
			ID:     xid.New().String(),
			Source: src,
		}
		for _, col := range maps {
			_, err := url.Parse(col.Dst)
			// If the URI is invalid, the results will be an empty string.
			if err != nil {
				col.Dst = ""
			}
			switch col.Arc {
			case config.SLOT_IA:
				doc.IA = col.Dst
			case config.SLOT_IS:
				doc.IS = col.Dst
			case config.SLOT_IP:
				doc.IP = col.Dst
			case config.SLOT_PH:
				doc.PH = col.Dst
			}
		}
		docs = append(docs, doc)
	}
	return
}

type collects map[string][]wayback.Collect

func groupBySrc(cols []wayback.Collect) collects {
	var c = make(collects)
	for _, col := range cols {
		c[col.Src] = append(c[col.Src], col)
	}
	return c
}

// Shutdown shuts down the Meilisearch publish service, it always return a nil error.
func (m *Meili) Shutdown() error {
	return nil
}
