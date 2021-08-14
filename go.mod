module github.com/wabarc/wayback

// +heroku goVersion go1.16

go 1.16

require (
	github.com/PuerkitoBio/goquery v1.7.1 // indirect
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/bwmarrin/discordgo v0.23.3-0.20210627161652-421e14965030
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/chromedp/cdproto v0.0.0-20210808225517-c36c1bd4c35e // indirect
	github.com/cretz/bine v0.2.0
	github.com/davecgh/go-spew v1.1.1
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/dghubble/oauth1 v0.7.0
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.12.0
	github.com/gabriel-vasile/mimetype v1.3.1
	github.com/go-shiori/go-readability v0.0.0-20210627123243-82cc33435520
	github.com/go-shiori/obelisk v0.0.0-20201115143556-8de0d40b0a9b // indirect
	github.com/google/go-github/v38 v38.1.0
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/iawia002/annie v0.0.0-20210720065824-ac5f36d99548
	github.com/klauspost/cpuid/v2 v2.0.8 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	// github.com/ipsn/go-libtor v1.0.329
	github.com/libp2p/go-libp2p-core v0.8.6 // indirect
	github.com/mattn/go-mastodon v0.0.5-0.20210515144304-86627ec7d635
	github.com/multiformats/go-multiaddr v0.3.3 // indirect
	github.com/multiformats/go-multihash v0.0.15 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.30.0
	github.com/robertkrimen/otto v0.0.0-20210614181706-373ff5438452 // indirect
	github.com/slack-go/slack v0.9.3
	github.com/spf13/cobra v1.2.1
	github.com/tdewolff/parse/v2 v2.5.19 // indirect
	github.com/thoj/go-ircevent v0.0.0-20190807115034-8e7ce4b5a1eb
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/wabarc/archive.is v1.3.0
	github.com/wabarc/archive.org v1.2.1-0.20210708220121-cb9b83ff9896
	github.com/wabarc/go-anonfile v0.1.0
	github.com/wabarc/go-catbox v0.1.0
	github.com/wabarc/helper v0.0.0-20210718171053-59c70d0b20c2
	github.com/wabarc/logger v0.0.0-20210730133522-86bd3f31e792
	github.com/wabarc/playback v0.0.0-20210718054702-cab6c6004933
	github.com/wabarc/screenshot v1.3.0
	github.com/wabarc/telegra.ph v0.0.0-20210718173528-f72fc9cdc365
	github.com/wabarc/warcraft v0.1.1-0.20210711171056-a5eec617b86c
	github.com/wabarc/wbipfs v0.2.0
	github.com/whyrusleeping/tar-utils v0.0.0-20201201191210-20a61371de5b // indirect
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/tucnak/telebot.v2 v2.4.0
	maunium.net/go/mautrix v0.9.17
	mvdan.cc/xurls/v2 v2.3.0 // indirect
)

replace github.com/go-shiori/obelisk => github.com/wabarc/obelisk v0.0.0-20210420023708-aac2bcc00a78

replace github.com/gabriel-vasile/mimetype => github.com/wabarc/mimetype v1.3.2-0.20210814000005-c49e03777d46
