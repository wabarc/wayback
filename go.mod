module github.com/wabarc/wayback

// +heroku goVersion go1.16

go 1.16

require (
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/cretz/bine v0.1.0
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/dghubble/oauth1 v0.7.0
	github.com/go-shiori/obelisk v0.0.0-20201115143556-8de0d40b0a9b // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-github/v33 v33.0.0
	github.com/gorilla/mux v1.8.0
	github.com/klauspost/cpuid/v2 v2.0.6 // indirect
	// github.com/ipsn/go-libtor v1.0.329
	github.com/libp2p/go-libp2p-core v0.8.5 // indirect
	github.com/mattn/go-mastodon v0.0.5-0.20210515144304-86627ec7d635
	github.com/multiformats/go-multiaddr v0.3.1 // indirect
	github.com/multiformats/go-multihash v0.0.15 // indirect
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/common v0.25.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/tdewolff/parse/v2 v2.5.15 // indirect
	github.com/thoj/go-ircevent v0.0.0-20190807115034-8e7ce4b5a1eb
	github.com/wabarc/archive.is v1.2.4-0.20210505135936-034a6c963560
	github.com/wabarc/archive.org v1.1.2
	github.com/wabarc/helper v0.0.0-20210511232523-5ac25c99226f
	github.com/wabarc/logger v0.0.0-20210417045349-d0d82e8e99ee
	github.com/wabarc/playback v0.0.0-20210418074547-4bf9d94a794d
	github.com/wabarc/screenshot v1.1.1 // indirect
	github.com/wabarc/telegra.ph v0.0.0-20210505140622-220623b0de58
	github.com/wabarc/wbipfs v0.1.3
	github.com/whyrusleeping/tar-utils v0.0.0-20201201191210-20a61371de5b // indirect
	go.etcd.io/bbolt v1.3.5
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/tucnak/telebot.v2 v2.3.5
	maunium.net/go/mautrix v0.9.12
)

replace github.com/go-shiori/obelisk => github.com/wabarc/obelisk v0.0.0-20210420023708-aac2bcc00a78
