module github.com/wabarc/wayback

go 1.13

require (
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/chromedp/chromedp v0.6.6 // indirect
	github.com/cretz/bine v0.1.0
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/dghubble/oauth1 v0.7.0
	github.com/go-shiori/obelisk v0.0.0-20201115143556-8de0d40b0a9b // indirect
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-github/v33 v33.0.0
	// github.com/ipsn/go-libtor v1.0.329
	github.com/libp2p/go-libp2p-core v0.8.5 // indirect
	github.com/magefile/mage v1.11.0 // indirect
	github.com/mattn/go-mastodon v0.0.4
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/multiformats/go-multiaddr v0.3.1 // indirect
	github.com/sirupsen/logrus v1.8.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/tdewolff/parse/v2 v2.5.11 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/wabarc/archive.is v1.1.2
	github.com/wabarc/archive.org v1.0.4
	github.com/wabarc/helper v0.0.0-20210227111116-538a9d8851db
	github.com/wabarc/telegra.ph v0.0.0-20210221072121-954117897e8e
	github.com/wabarc/wbipfs v0.1.2
	github.com/whyrusleeping/tar-utils v0.0.0-20201201191210-20a61371de5b // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210227040730-b0d1d43c014d // indirect
)

replace github.com/go-shiori/obelisk => github.com/wabarc/obelisk v0.0.0-20201203134337-61a4e7973e4c

replace github.com/mattn/go-mastodon => github.com/wabarc/go-mastodon v0.0.5-0.20210225142645-546cc5f43352
