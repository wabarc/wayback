// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package reduxer // import "github.com/wabarc/wayback/reduxer"

import (
	"github.com/go-shiori/go-readability"
	"github.com/wabarc/screenshot"
)

func BundleExample() Reduxer {
	rdx := NewReduxer()
	bnd := &bundle{
		artifact: Artifact{
			Img: Asset{
				Local: "/path/to/image",
				Remote: Remote{
					Anonfile: "https://anonfiles.com/FbZfSa9eu4",
					Catbox:   "https://files.catbox.moe/9u6yvu.png",
				},
			},
			PDF: Asset{
				Local: "/path/to/pdf",
				Remote: Remote{
					Anonfile: "https://anonfiles.com/r4G8Sb90ud",
					Catbox:   "https://files.catbox.moe/q73uqh.pdf",
				},
			},
			Raw: Asset{
				Local: "/path/to/htm",
				Remote: Remote{
					Anonfile: "https://anonfiles.com/pbG4Se94ua",
					Catbox:   "https://files.catbox.moe/bph1g6.htm",
				},
			},
			Txt: Asset{
				Local: "/path/to/txt",
				Remote: Remote{
					Anonfile: "https://anonfiles.com/naG6S09bu1",
					Catbox:   "https://files.catbox.moe/wwrby6.txt",
				},
			},
			HAR: Asset{
				Local: "/path/to/har",
				Remote: Remote{
					Anonfile: "https://anonfiles.com/n1paZfB3ub",
					Catbox:   "https://files.catbox.moe/3agtva.har",
				},
			},
			HTM: Asset{
				Local: "/path/to/single-htm",
				Remote: Remote{
					Anonfile: "https://anonfiles.com/v4G4S09abc",
					Catbox:   "",
				},
			},
			WARC: Asset{
				Local: "/path/to/warc",
				Remote: Remote{
					Anonfile: "https://anonfiles.com/v4G4S09auc",
					Catbox:   "invalid-url-moe/kkai0w.warc",
				},
			},
			Media: Asset{
				Local: "",
				Remote: Remote{
					Anonfile: "",
					Catbox:   "",
				},
			},
		},
		shots: &screenshot.Screenshots[screenshot.Path]{
			Title: "Example",
		},
		article: readability.Article{
			Content: `<!doctype html>
<html>
<head>
    <title>Example Domain</title>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is for use in illustrative examples in documents. You may use this
    domain in literature without prior coordination or asking for permission.</p>
    <p><a href="https://www.iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>`,
			TextContent: `This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.

More information...`,
		},
	}

	rdx.Store(Src("https://example.com/"), bnd)

	return rdx
}
