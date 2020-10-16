# Wayback

[![Go Report Card](https://goreportcard.com/badge/github.com/wabarc/wayback)](https://goreportcard.com/report/github.com/wabarc/wayback)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/wabarc/wayback/Go?color=brightgreen)](https://github.com/wabarc/wayback/actions)
[![Releases](https://img.shields.io/github/v/release/wabarc/wayback.svg?include_prereleases&color=blue)](https://github.com/wabarc/wayback/releases)
[![LICENSE](https://img.shields.io/github/license/wabarc/wayback.svg?color=green)](https://github.com/wabarc/wayback/blob/master/LICENSE)
[![Docker Automated build](https://img.shields.io/docker/automated/wabarc/wayback)](https://hub.docker.com/r/wabarc/wayback)
[![wayback](https://snapcraft.io/wayback/badge.svg)](https://snapcraft.io/wayback)

`wabarc/wayback` is a tool that supports running as a command-line tool and docker container, purpose to snapshot webpage to time capsules.

## Installation

```sh
$ go get -u github.com/wabarc/wayback/cmd/wayback
```

Using [Snapcraft](https://snapcraft.io/wayback) (on GNU/Linux)

```sh
$ sudo snap install wayback
```

## Usage

- Running as CLI command or Docker container
- Running with telegram bot

### Command line

```sh
$ wayback -h
A CLI tool for wayback webpages.

Usage:
  wayback [flags]

Examples:
  wayback https://www.wikipedia.org
  wayback https://www.fsf.org https://www.eff.org
  wayback --ia https://www.fsf.org
  wayback --ip https://www.fsf.org
  wayback --ia --is -d telegram -t your-telegram-bot-token
  WAYBACK_SLOT=pinata WAYBACK_APIKEY=YOUR-PINATA-APIKEY \
    WAYBACK_SECRET=YOUR-PINATA-SECRET wayback --ip https://www.fsf.org

Flags:
  -c, --chatid string      Channel ID. default: ""
  -d, --daemon string      Run as daemon service.
      --debug              Enable debug mode. default: false
  -h, --help               help for wayback
      --ia                 Wayback webpages to Internet Archive.
      --ip                 Wayback webpages to IPFS. (default false)
      --ipfs-host string   IPFS daemon host, do not require, unless enable ipfs. (default "127.0.0.1")
  -m, --ipfs-mode string   IPFS mode. (default "pinner")
  -p, --ipfs-port uint     IPFS daemon port. (default 5001)
      --is                 Wayback webpages to Archive Today.
  -t, --token string       Telegram bot API Token, required.
      --tor                Snapshot webpage use tor proxy.
  -v, --version            version for wayback
```

#### Examples

Wayback one or more url to *Internet Archive* **and** *archive.today*:

```sh
$ wayback https://www.wikipedia.org

$ wayback https://www.fsf.org https://www.eff.org
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

For the IPFS, also can use a specific pinner service:

```sh
$ export WAYBACK_SLOT=pinata
$ export WAYBACK_APIKEY=YOUR-PINATA-APIKEY
$ export WAYBACK_SECRET=YOUR-PINATA-SECRET
$ wayback --ip https://www.fsf.org

// or

$ WAYBACK_SLOT=pinata WAYBACK_APIKEY=YOUR-PINATA-APIKEY \
$ WAYBACK_SECRET=YOUR-PINATA-SECRET wayback --ip https://www.fsf.org
```

TIP: [more details](https://github.com/wabarc/ipfs-pinner) about pinner service.

With telegram bot:

```sh
$ wayback --ia --is --ip -d telegram -t your-telegram-bot-token
```

Publish message to your Telegram channel at the same time:

```sh
$ wayback --ia --is --ip -d telegram -t your-telegram-bot-token -c your-telegram-channel-name
```

Also can run with debug mode:

```sh
$ wayback -d telegram -t YOUR-BOT-TOKEN --debug
```

### Docker/Podman

```sh
$ docker pull wabarc/wayback
$ docker run -d wabarc/wayback -d telegram -t YOUR-BOT-TOKEN # without telegram channel
$ docker run -d wabarc/wayback -d telegram -t YOUR-BOT-TOKEN -c YOUR-CHANNEL-USERNAME # with telegram channel
```

## Deploy on Heroku

See: [wabarc/on-heroku](https://github.com/wabarc/on-heroku)

## TODO

[Archive.org](https://web.archive.org/) and [Archive.today](https://archive.today/) are currently supported, the next step mind support the followings platform:

- [x] [IPFS](https://ipfs.io/)
- [ ] [ZeroNet](https://zeronet.io/)

## Telegram bot

- [Bots: An introduction for developers](https://core.telegram.org/bots)
- [How do I create a bot?](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- [An example bot](http://t.me/wabarc_bot)
- [An example channel](http://t.me/wabarc)

## Related projects

- [duty-machine](https://github.com/duty-machine/duty-machine)
- [ipfs-pinner](https://github.com/wabarc/ipfs-pinner)
- [on-heroku](https://github.com/wabarc/on-heroku)

## License

This software is released under the terms of the GNU General Public License v3.0. See the [LICENSE](https://github.com/wabarc/wayback/blob/master/LICENSE) file for details.
