// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package service // import "github.com/wabarc/wayback/service"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/wabarc/wayback"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
)

var (
	timeout    = 10 * time.Second
	indexing   = `capsules`
	primaryKey = `id`

	userAgent   = `WaybackArchiver/1.0`
	contentType = `application/json`

	errIndexNotFound = errors.New(fmt.Sprintf(`indexing %s not found`, indexing))
	errIndexNotMatch = errors.New(fmt.Sprintf(`indexing %s not match`, indexing))

	// TODO: find a better way to handle it.
	meili *Meili
)

// Meili represents a Meilisearch client.
type Meili struct {
	// Meilisearch server API endpoint.
	endpoint string

	// Meilisearch indexing name.
	indexing string

	// Meilisearch admin API key, which can be emptied.
	apikey string

	client *http.Client
}

// NewMeili returns a Meilisearch client.
func NewMeili(endpoint, apikey, idxname string) *Meili {
	client := &http.Client{
		Timeout: timeout,
	}
	if idxname == "" {
		idxname = indexing
	}
	meili = &Meili{
		endpoint: endpoint,
		indexing: idxname,
		apikey:   apikey,
		client:   client,
	}

	return meili
}

// Setup creates an index if one does not already exist on the server.
func (m *Meili) Setup() error {
	err := m.existIndex()
	if errors.Is(err, errIndexNotFound) {
		err = m.createIndex()
	}
	if err != nil {
		return err
	}
	return m.sortable()
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

	body, err := ioutil.ReadAll(resp.Body)
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
	UID        int       `json:"uid"`
	IndexUID   string    `json:"indexUid"`
	Status     string    `json:"status"`
	Type       string    `json:"type"`
	EnqueuedAt time.Time `json:"enqueuedAt"`
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

	body, err := ioutil.ReadAll(resp.Body)
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
	payload := fmt.Sprintf(`["id"]`)
	resp, err := m.do(http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return errors.Wrap(err, `set sortable attributes: request failed`)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(`set sortable attributes: unexpected status: ` + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
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
func (m *Meili) push(cols []wayback.Collect) error {
	if len(cols) == 0 {
		return errors.New(`push documents failed: cols empty`)
	}

	buf, err := json.Marshal(m.documents(cols))
	if err != nil {
		return errors.Wrap(err, `push document: marshal docs failed`)
	}

	endpoint := fmt.Sprintf(`%s/indexes/%s/documents`, m.endpoint, m.indexing)
	resp, err := m.do(http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return errors.Wrap(err, `push document: failed`)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(`push document: unexpected status: ` + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `push document: reads body failed`)
	}

	var idx creates
	if err := json.Unmarshal(body, &idx); err != nil {
		return errors.Wrap(err, `push document: unmarshal json failed`)
	}
	if idx.IndexUID != m.indexing {
		return errIndexNotMatch
	}

	return nil
}

func (m *Meili) do(method, url string, body io.Reader) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Authorization", "Bearer "+m.apikey)
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
