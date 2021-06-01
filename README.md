# Wayback

[![LICENSE](https://img.shields.io/github/license/wabarc/wayback.svg?color=green)](https://github.com/wabarc/wayback/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/wabarc/wayback)](https://goreportcard.com/report/github.com/wabarc/wayback)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/wabarc/wayback/Go?color=brightgreen)](https://github.com/wabarc/wayback/actions)
[![Go Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/wabarc/wayback)
[![Releases](https://img.shields.io/github/v/release/wabarc/wayback.svg?include_prereleases&color=blue)](https://github.com/wabarc/wayback/releases)

[![Telegram Bot](https://img.shields.io/badge/Telegram-bot-3dbeff.svg)](https://t.me/wabarc_bot)
[![Telegram Channel](https://img.shields.io/badge/Telegram-channel-3dbeff.svg)](https://t.me/wabarc)
[![Matrix Bot](https://img.shields.io/badge/Matrix-bot-0a976f.svg)](https://matrix.to/#/@wabarc_bot:matrix.org)
[![Matrix Room](https://img.shields.io/badge/Matrix-room-0a976f.svg)](https://matrix.to/#/#wabarc:matrix.org)
[![Tor Hidden Service](https://img.shields.io/badge/Tor%20Hidden%20Service-472756.svg)](http://wizmoki7pm5r2bco4holq467cq53kicttzge47fmxtis4x6tpt2u4nqd.onion/)
[![World Wide Web](https://img.shields.io/badge/Web-15aabf.svg)](https://initium.eu.org/)

Wayback is a tool that supports running as a command-line tool and docker container, purpose to snapshot webpage to time capsules.

## Features

- Cross platform
- Batch wayback URLs
- Builtin CLI (`wayback`)
- Serve as Tor Hidden Service or local web entry
- Wayback to Internet Archive, archive.today, IPFS and Telegraph easier
- Interactive with IRC, Martix, Telegram bot, Mastodon and Twitter as daemon service
- Support publish wayback results to Telegram channel, Mastodon and GitHub Issues

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

Via [APT](https://github.com/wabarc/apt-repo):

```bash
$ curl -s https://apt.wabarc.eu.org/KEY.gpg | sudo apt-key add -
$ sudo echo "deb https://apt.wabarc.eu.org/ /" > /etc/apt/sources.list.d/wayback.list
$ sudo apt update
$ sudo apt install wayback
```

Via [RPM](https://github.com/wabarc/rpm-repo):

```
$ sudo cat > /etc/yum.repos.d/wayback.repo<< EOF
[wayback]
name=Wayback Repository
baseurl=https://rpm.wabarc.eu.org/x86_64/
enabled=1
gpgcheck=0
EOF

$ sudo yum install -y wayback
```

Via [Homebrew](https://github.com/wabarc/homebrew-wayback):

```
$ brew tap wabarc/wayback
$ brew install wayback
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
      --chatid string      Telegram channel id.
  -c, --config string      Configuration file path, defaults: ./wayback.conf, ~/wayback.conf, /etc/wayback.conf
  -d, --daemon strings     Run as daemon service, supported services are telegram, web, mastodon, twitter, irc
      --debug              Enable debug mode. (default false)
  -h, --help               help for wayback
      --ia                 Wayback webpages to Internet Archive.
      --info               Show application information.
      --ip                 Wayback webpages to IPFS. (default false)
      --ipfs-host string   IPFS daemon host, do not require, unless enable ipfs. (default "127.0.0.1")
  -m, --ipfs-mode string   IPFS mode. (default "pinner")
  -p, --ipfs-port uint     IPFS daemon port. (default 5001)
      --is                 Wayback webpages to Archive Today.
      --ph                 Wayback webpages to Telegraph. (default false)
      --print              Show application configurations.
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

With telegram bot:

```sh
$ wayback --ia --is --ip -d telegram -t your-telegram-bot-token
```

Publish message to your Telegram channel at the same time:

```sh
$ wayback --ia --is --ip -d telegram -t your-telegram-bot-token --chatid your-telegram-channel-name
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

By default, `wayback` looks for configuration options from this files, the following are parsed:

- `./wayback.conf`
- `~/wayback.conf`
- `/etc/wayback.conf`

Use the `-c` / `--config` option to specify the build definition file to use.

You can also specify configuration options either via command flags or via environment variables, an overview of all options below.

| Flags               | Environment Variable              | Default                 | Description                                                  |
| ------------------- | --------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `--debug`           | `DEBUG`                           | `false`                 | Enable debug mode                                            |
| `-c`, `--config`    | -                                 | -                       | Configuration file path, defaults: `./wayback.conf`, `~/wayback.conf`, `/etc/wayback.conf` |
| -                   | `LOG_TIME`                        | `true`                  | Display the date and time in log messages                    |
| -                   | `ENABLE_METRICS`                  | `false`                 | Enable metrics collector                                     |
| -                   | `CHROME_REMOTE_ADDR`              | -                       | Chrome/Chromium remote debugging address, for screenshot     |
| -                   | `WAYBACK_POOLING_SIZE`            | `3`                     | Number of worker pool for wayback at once                    |
| -                   | `WAYBACK_BOLT_PATH`               | `./wayback.db`          | File path of bolt database                                   |
| `-d`, `--daemon`    | -                                 | -                       | Run as daemon service, e.g. `telegram`, `web`, `mastodon`, `twitter` |
| `--ia`              | `WAYBACK_ENABLE_IA`               | `true`                  | Wayback webpages to **Internet Archive**                     |
| `--is`              | `WAYBACK_ENABLE_IS`               | `true`                  | Wayback webpages to **Archive Today**                        |
| `--ip`              | `WAYBACK_ENABLE_IP`               | `false`                 | Wayback webpages to **IPFS**                                 |
| `--ph`              | `WAYBACK_ENABLE_PH`               | `false`                 | Wayback webpages to **[Telegra.ph](https://telegra.ph)**, required Chrome/Chromium |
| `--ipfs-host`       | `WAYBACK_IPFS_HOST`               | `127.0.0.1`             | IPFS daemon service host                                     |
| `-p`, `--ipfs-port` | `WAYBACK_IPFS_PORT`               | `5001`                  | IPFS daemon service port                                     |
| `-m`, `--ipfs-mode` | `WAYBACK_IPFS_MODE`               | `pinner`                | IPFS mode for preserve webpage, e.g. `daemon`, `pinner`      |
| -                   | `WAYBACK_GITHUB_TOKEN`            | -                       | GitHub Personal Access Token, required the `repo` scope      |
| -                   | `WAYBACK_GITHUB_OWNER`            | -                       | GitHub account name                                          |
| -                   | `WAYBACK_GITHUB_REPO`             | -                       | GitHub repository to publish results                         |
| `-t`, `--token`     | `WAYBACK_TELEGRAM_TOKEN`          | -                       | Telegram Bot API Token                                       |
| `--chatid`          | `WAYBACK_TELEGRAM_CHANNEL`        | -                       | The **Telegram Channel** name for publish archived result    |
| -                   | `WAYBACK_TELEGRAM_HELPTEXT`       | -                       | The help text for Telegram bot command                       |
| -                   | `WAYBACK_MASTODON_SERVER`         | -                       | Domain of Mastodon instance                                  |
| -                   | `WAYBACK_MASTODON_KEY`            | -                       | The client key of your Mastodon application                  |
| -                   | `WAYBACK_MASTODON_SECRET`         | -                       | The client secret of your Mastodon application               |
| -                   | `WAYBACK_MASTODON_TOKEN`          | -                       | The access token of your Mastodon application                |
| -                   | `WAYBACK_TWITTER_CONSUMER_KEY`    | -                       | The customer key of your Twitter application                 |
| -                   | `WAYBACK_TWITTER_CONSUMER_SECRET` | -                       | The customer secret of your Twitter application              |
| -                   | `WAYBACK_TWITTER_ACCESS_TOKEN`    | -                       | The access token of your Twitter application                 |
| -                   | `WAYBACK_TWITTER_ACCESS_SECRET`   | -                       | The access secret of your Twitter application                |
| -                   | `WAYBACK_IRC_NICK`                | -                       | IRC nick                                                     |
| -                   | `WAYBACK_IRC_PASSWORD`            | -                       | IRC password                                                 |
| -                   | `WAYBACK_IRC_CHANNEL`             | -                       | IRC channel                                                  |
| -                   | `WAYBACK_IRC_SERVER`              | `irc.libera.chat:6697`  | IRC server, required TLS                                     |
| -                   | `WAYBACK_MATRIX_HOMESERVER`       | `https://matrix.org`    | Matrix homeserver                                            |
| -                   | `WAYBACK_MATRIX_USERID`           | -                       | Matrix unique user ID, format: `@foo:example.com`            |
| -                   | `WAYBACK_MATRIX_ROOMID`           | -                       | Matrix internal room ID, format: `!bar:example.com`          |
| -                   | `WAYBACK_MATRIX_PASSWORD`         | -                       | Matrix password                                              |
| `--tor`             | `WAYBACK_USE_TOR`                 | `false`                 | Snapshot webpage via Tor anonymity network                   |
| `--tor-key`         | `WAYBACK_TOR_PRIVKEY`             | -                       | The private key for Tor Hidden Service                       |
| -                   | `WAYBACK_TOR_LOCAL_PORT`          | `8964`                  | Local port for Tor Hidden Service, also support for a **reverse proxy** |
| -                   | `WAYBACK_TOR_REMOTE_PORTS`        | `80`                    | Remote ports for Tor Hidden Service, e.g. `WAYBACK_TOR_REMOTE_PORTS=80,81` |
| -                   | `WAYBACK_TORRC`                   | `/etc/tor/torrc`        | Using `torrc` for Tor Hidden Service                         |
| -                   | `WAYBACK_SLOT`                    | -                       | Pinning service for IPFS mode of pinner, see [ipfs-pinner](https://github.com/wabarc/ipfs-pinner#supported-pinning-services) |
| -                   | `WAYBACK_APIKEY`                  | -                       | API key for pinning service                                  |
| -                   | `WAYBACK_SECRET`                  | -                       | API secret for pinning service                               |

If both of the definition file and environment variables are specified, they are all will be read and apply,
and preferred from the environment variable for the same item.

Prints the resulting options of the targets with `--print`, in a Go struct with type, without running the `wayback`.

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

## Services

### Telegram bot

- [Bots: An introduction for developers](https://core.telegram.org/bots)
- [How do I create a bot?](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- [An example bot](http://t.me/wabarc_bot)
- [An example channel](http://t.me/wabarc)

### Mastodon bot

Bot friendly instance:

- [botsin.space](https://botsin.space/about/more)

## F.A.Q

**Q: How to keep the Tor hidden service hostname?**

A: For the first time to run the `wayback` service, keep the key from the output message (the key is the part after `private key:` below) 
and next time to run the `wayback` service to place the key to the `--tor-key` option or the `WAYBACK_TOR_PRIVKEY` environment variable.
```
[INFO] Web: important to keep the private key: d005473a611d2b23e54d6446dfe209cb6c52ddd698818d1233b1d750f790445fcfb5ece556fe5ee3b4724ac6bea7431898ee788c6011febba7f779c85845ae87
```

## Contributing

We encourage all contributions to this repository! Open an issue! Or open a Pull Request!

If you're interested in contributing to `wayback` itself, read our [contributing guide](./CONTRIBUTING.md) to get started.

Note: All interaction here should conform to the [Code of Conduct](./CODE_OF_CONDUCT.md).

## License

This software is released under the terms of the GNU General Public License v3.0. See the [LICENSE](https://github.com/wabarc/wayback/blob/main/LICENSE) file for details.
