# Wayback

[![LICENSE](https://img.shields.io/github/license/wabarc/wayback.svg?color=green)](https://github.com/wabarc/wayback/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/wabarc/wayback)](https://goreportcard.com/report/github.com/wabarc/wayback)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/wabarc/wayback/Go?color=brightgreen)](https://github.com/wabarc/wayback/actions)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/wabarc/wayback)
[![Releases](https://img.shields.io/github/v/release/wabarc/wayback.svg?include_prereleases&color=blue)](https://github.com/wabarc/wayback/releases)
[![Docker Automated build](https://github.com/wabarc/wayback/workflows/Docker/badge.svg)](https://hub.docker.com/r/wabarc/wayback)
[![Snapcraft](https://github.com/wabarc/wayback/workflows/Snapcraft/badge.svg)](https://snapcraft.io/wayback)

Wayback is a tool that supports running as a command-line tool and docker container, purpose to snapshot webpage to time capsules.

## Feature

- CLI tool
- Interactive with Telegram bot
- Serve as Tor Hidden Service or local web entry
- Wayback to Internet Archive, archive.today, IPFS, etc

## Installation

From source:

```sh
$ go get -u github.com/wabarc/wayback/cmd/wayback
```

From [GoBinaries](https://gobinaries.com/):

```sh
$ curl -sf https://gobinaries.com/wabarc/wayback/cmd/wayback | sh
```

Using [Snapcraft](https://snapcraft.io/wayback) (on GNU/Linux)

```sh
$ sudo snap install wayback
```

See more on [releases](https://github.com/wabarc/wayback/releases).

## Usage

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
  wayback --ia --is -d telegram -t your-telegram-bot-token
  WAYBACK_SLOT=pinata WAYBACK_APIKEY=YOUR-PINATA-APIKEY \
    WAYBACK_SECRET=YOUR-PINATA-SECRET wayback --ip https://www.fsf.org

Flags:
  -c, --chatid string      Telegram channel id.
  -d, --daemon strings     Run as daemon service, e.g. telegram, web
      --debug              Enable debug mode. (default false)
  -h, --help               help for wayback
      --ia                 Wayback webpages to Internet Archive.
      --ip                 Wayback webpages to IPFS. (default false)
      --ipfs-host string   IPFS daemon host, do not require, unless enable ipfs. (default "127.0.0.1")
  -m, --ipfs-mode string   IPFS mode. (default "pinner")
  -p, --ipfs-port uint     IPFS daemon port. (default 5001)
      --is                 Wayback webpages to Archive Today.
      --ph                 Wayback webpages to Telegraph. (default false)
  -t, --token string       Telegram Bot API Token.
      --tor                Snapshot webpage via Tor anonymity network.
      --tor-key string     The private key for Tor Hidden Service.
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

Both serve on Telegram and Tor hidden service:

```sh
$ wayback -d telegram -t YOUT-BOT-TOKEN -d web
```

#### Configuration Parameters

You can specify configuration options either via command flags or via environment variables, an overview of all options below.


| Flags               | Environment Variable       | Default     | Description                                                  |
| ------------------- | -------------------------- | ----------- | ------------------------------------------------------------ |
| `--debug`           | `DEBUG`                    | `false`     | Enable debug mode                                            |
| -                   | `LOG_TIME`                 | `true`      | Display the date and time in log messages                    |
| `-d`, `--daemon`    | -                          | -           | Run as daemon service, e.g. `telegram`, `web`                |
| `--ia`              | `WAYBACK_ENABLE_IA`        | `true`      | Wayback webpages to **Internet Archive**                     |
| `--is`              | `WAYBACK_ENABLE_IS`        | `true`      | Wayback webpages to **Archive Today**                        |
| `--ip`              | `WAYBACK_ENABLE_IP`        | `false`     | Wayback webpages to **IPFS**                                 |
| `--ph`              | `WAYBACK_ENABLE_PH`        | `false`     | Wayback webpages to **[Telegra.ph](https://telegra.ph)**, required Chrome/Chromium |
| `--ipfs-host`       | `WAYBACK_IPFS_HOST`        | `127.0.0.1` | IPFS daemon service host                                     |
| `-p`, `--ipfs-port` | `WAYBACK_IPFS_PORT`        | `5001`      | IPFS daemon service port                                     |
| `-m`, `--ipfs-mode` | `WAYBACK_IPFS_MODE`        | `pinner`    | IPFS mode for preserve webpage, e.g. `daemon`, `pinner`      |
| `-t`, `--token`     | `WAYBACK_TELEGRAM_TOKEN`   | -           | Telegram Bot API Token                                       |
| `-c`, `--chatid`    | `WAYBACK_TELEGRAM_CHANNEL` | -           | The **Telegram Channel** name for publish archived result    |
| `--tor`             | `WAYBACK_USE_TOR`          | `false`     | Snapshot webpage via Tor anonymity network                   |
| `--tor-key`         | `WAYBACK_TOR_PRIVKEY`      | -           | The private key for Tor Hidden Service                       |
| -                   | `WAYBACK_TOR_LOCAL_PORT`   | -           | Local port for Tor Hidden Service, also support for a **reverse proxy** |
| -                   | `WAYBACK_TOR_REMOTE_PORTS` | `80`        | Remote ports for Tor Hidden Service, e.g. `WAYBACK_TOR_REMOTE_PORTS=80,81` |
| -                   | `WAYBACK_TORRC`            | `/etc/tor/torrc` | Using `torrc` for Tor Hidden Service |

### Docker/Podman

```sh
$ docker pull wabarc/wayback
$ docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN # without telegram channel
$ docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN -c YOUR-CHANNEL-USERNAME # with telegram channel
```

## Deployment

- [wabarc/on-heroku](https://github.com/wabarc/on-heroku)
- [wabarc/on-github](https://github.com/wabarc/on-github)

## TODO

[Archive.org](https://web.archive.org/) and [Archive.today](https://archive.today/) are currently supported, the next step mind support the followings platform:

- [x] [IPFS](https://ipfs.io/)
- [ ] ~~[ZeroNet](https://zeronet.io/)~~

## Telegram bot

- [Bots: An introduction for developers](https://core.telegram.org/bots)
- [How do I create a bot?](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- [An example bot](http://t.me/wabarc_bot)
- [An example channel](http://t.me/wabarc)

## F.A.Q

**Q: How to keep the Tor hidden service hostname?**

A: For the first time to run the `wayback` service, keep the key from the output message (the key is the part after `private key:` below) 
and next time to run the `wayback` service to place the key to the `--tor-key` option or the `WAYBACK_TOR_PRIVKEY` environment variable.
```
[INFO] Web: important to keep the private key: d005473a611d2b23e54d6446dfe209cb6c52ddd698818d1233b1d750f790445fcfb5ece556fe5ee3b4724ac6bea7431898ee788c6011febba7f779c85845ae87
```

## License

This software is released under the terms of the GNU General Public License v3.0. See the [LICENSE](https://github.com/wabarc/wayback/blob/main/LICENSE) file for details.
