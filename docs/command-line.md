## Command line

```sh
$ wayback -h

A command-line tool and daemon service for archiving webpages.

Usage:
  wayback [flags]

Examples:
  wayback https://www.wikipedia.org
  wayback https://www.fsf.org https://www.eff.org
  wayback --ia https://www.fsf.org
  wayback --ia --is -d telegram -t your-telegram-bot-token
  WAYBACK_SLOT=pinata WAYBACK_APIKEY=YOUR-PINATA-APIKEY \
    WAYBACK_SECRET=YOUR-PINATA-SECRET wayback --ip https://www.fsf.org

Flags:
      --chatid string      Telegram channel id
  -c, --config string      Configuration file path, defaults: ./wayback.conf, ~/wayback.conf, /etc/wayback.conf
  -d, --daemon strings     Run as daemon service, supported services are telegram, web, mastodon, twitter, discord, slack, irc
      --debug              Enable debug mode (default mode is false)
  -h, --help               help for wayback
      --ia                 Wayback webpages to Internet Archive
      --info               Show application information
      --ip                 Wayback webpages to IPFS
      --ipfs-host string   IPFS daemon host, do not require, unless enable ipfs (default "127.0.0.1")
  -m, --ipfs-mode string   IPFS mode (default "pinner")
  -p, --ipfs-port uint     IPFS daemon port (default 5001)
      --is                 Wayback webpages to Archive Today
      --ph                 Wayback webpages to Telegraph
      --print              Show application configurations
  -t, --token string       Telegram Bot API Token
      --tor                Snapshot webpage via Tor anonymity network
      --tor-key string     The private key for Tor Hidden Service
  -v, --version            version for wayback
```

## Examples

Wayback one or more url to *Internet Archive* **and** *archive.today*:

```sh
wayback https://www.wikipedia.org

wayback https://www.fsf.org https://www.eff.org
```

Wayback url to *Internet Archive* **or** *archive.today* **or** *IPFS*:

```sh
// Internet Archive
$ wayback --ia https://www.fsf.org

// archive.today
$ wayback --is https://www.fsf.org

// IPFS
$ wayback --ip https://www.fsf.org
```

For using IPFS, also can specify a pinning service:

```sh
$ export WAYBACK_SLOT=pinata
$ export WAYBACK_APIKEY=YOUR-PINATA-APIKEY
$ export WAYBACK_SECRET=YOUR-PINATA-SECRET
$ wayback --ip https://www.fsf.org

// or

$ WAYBACK_SLOT=pinata WAYBACK_APIKEY=YOUR-PINATA-APIKEY \
$ WAYBACK_SECRET=YOUR-PINATA-SECRET wayback --ip https://www.fsf.org
```

More details about [pinning service](https://github.com/wabarc/ipfs-pinner).

Wayback URLs from file:

```sh
wayback url.txt
```

With redirection:

```sh
cat url.txt | wayback
```
