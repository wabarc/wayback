# Wayback

[![LICENSE](https://img.shields.io/github/license/wabarc/wayback.svg?color=green)](https://github.com/wabarc/wayback/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/wabarc/wayback)](https://goreportcard.com/report/github.com/wabarc/wayback)
[![Test Coverage](https://codecov.io/gh/wabarc/wayback/branch/main/graph/badge.svg)](https://codecov.io/gh/wabarc/wayback)
[![Go Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/wabarc/wayback)
[![Releases](https://img.shields.io/github/v/release/wabarc/wayback.svg?include_prereleases&color=blue)](https://github.com/wabarc/wayback/releases)

[![Telegram Bot](https://img.shields.io/badge/Telegram-bot-3dbeff.svg)](https://t.me/wabarc_bot)
[![Discord Bot](https://img.shields.io/badge/Discord-bot-3dbeff.svg)](https://discord.com/api/oauth2/authorize?client_id=863324809206169640&permissions=2147796992&scope=bot%20applications.commands)
[![Matrix Bot](https://img.shields.io/badge/Matrix-bot-0a976f.svg)](https://matrix.to/#/@wabarc_bot:matrix.org)
[![Matrix Room](https://img.shields.io/badge/Matrix-room-0a976f.svg)](https://matrix.to/#/#wabarc:matrix.org)
[![Tor Hidden Service](https://img.shields.io/badge/Tor%20Hidden%20Service-472756.svg)](http://wabarcoww2bxmdbixj7sjwggv3fonh2rpflfiildegcydk5udkdckdyd.onion/)
[![World Wide Web](https://img.shields.io/badge/Web-15aabf.svg)](https://initium.eu.org/)
[![Nostr](https://img.shields.io/badge/Nostr-8e44ad.svg)](https://iris.to/#/profile/npub1gm4xeu8wlt6aa56zenutkwa0ppjng5axsscv424d0xvv5jalxxzs4hjukz)

Wayback is a tool that supports running as a command-line tool and docker container, purpose to snapshot webpage to time capsules.

Supported Golang version: See [.github/workflows/testing.yml](./.github/workflows/testing.yml)

## Features

- Free and open-source
- Cross-platform compatibility
- Batch wayback URLs for faster archiving
- Built-in CLI (`wayback`) for easy use
- Serve as a Tor Hidden Service or local web entry for added privacy and accessibility
- Easier wayback to Internet Archive, archive.today, IPFS and Telegraph integration
- Interactive with IRC, Matrix, Telegram bot, Discord bot, Mastodon, and Twitter as a daemon service for convenient use
- Supports publishing wayback results to Telegram channel, Mastodon, and GitHub Issues for easy sharing
- Supports storing archived files to disk for offline use
- Download streaming media (requires [FFmpeg](https://ffmpeg.org/)) for convenient media archiving.

## Installation

The simplest, cross-platform way is to download from [GitHub Releases](https://github.com/wabarc/wayback/releases) and place the executable file in your PATH.

From source:

```sh
go install github.com/wabarc/wayback/cmd/wayback@latest
```

From GitHub Releases:

```sh
curl -fsSL https://github.com/wabarc/wayback/raw/main/install.sh | sh
```

or via [Bina](https://bina.egoist.dev/):

```sh
curl -fsSL https://bina.egoist.dev/wabarc/wayback | sh
```

Using [Snapcraft](https://snapcraft.io/wayback) (on GNU/Linux)

```sh
sudo snap install wayback
```

Via [APT](https://repo.wabarc.eu.org/deb:wayback):

```bash
curl -fsSL https://repo.wabarc.eu.org/apt/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/packages.wabarc.gpg
echo "deb [arch=amd64,arm64,armhf signed-by=/usr/share/keyrings/packages.wabarc.gpg] https://repo.wabarc.eu.org/apt/ /" | sudo tee /etc/apt/sources.list.d/wayback.list
sudo apt update
sudo apt install wayback
```

Via [RPM](https://repo.wabarc.eu.org/rpm:wayback):

```bash
sudo rpm --import https://repo.wabarc.eu.org/yum/gpg.key
sudo tee /etc/yum.repos.d/wayback.repo > /dev/null <<EOT
[wayback]
name=Wayback Archiver
baseurl=https://repo.wabarc.eu.org/yum/
enabled=1
gpgcheck=1
gpgkey=https://repo.wabarc.eu.org/yum/gpg.key
EOT

sudo dnf install -y wayback
```

Via [Homebrew](https://github.com/wabarc/homebrew-wayback):

```shell
brew tap wabarc/wayback
brew install wayback
```

## Usage

### Command line

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

#### Examples

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

With telegram bot:

```sh
wayback --ia --is --ip -d telegram -t your-telegram-bot-token
```

Publish message to your Telegram channel at the same time:

```sh
wayback --ia --is --ip -d telegram -t your-telegram-bot-token --chatid your-telegram-channel-name
```

Also can run with debug mode:

```sh
wayback -d telegram -t YOUR-BOT-TOKEN --debug
```

Both serve on Telegram and Tor hidden service:

```sh
wayback -d telegram -t YOUT-BOT-TOKEN -d web
```

URLs from file:

```sh
wayback url.txt
```

```sh
cat url.txt | wayback
```

#### Configuration Parameters

By default, `wayback` looks for configuration options from this files, the following are parsed:

- `./wayback.conf`
- `~/wayback.conf`
- `/etc/wayback.conf`

Use the `-c` / `--config` option to specify the build definition file to use.

You can also specify configuration options either via command flags or via environment variables, an overview of all options below.

| Flags               | Environment Variable              | Default                    | Description                                                  |
| ------------------- | --------------------------------- | -------------------------- | ------------------------------------------------------------ |
| `--debug`           | `DEBUG`                           | `false`                    | Enable debug mode, override `LOG_LEVEL`                      |
| `-c`, `--config`    | -                                 | -                          | Configuration file path, defaults: `./wayback.conf`, `~/wayback.conf`, `/etc/wayback.conf` |
| -                   | `LOG_TIME`                        | `true`                     | Display the date and time in log messages                    |
| -                   | `LOG_LEVEL`                       | `info`                     | Log level, supported level are `debug`, `info`, `warn`, `error`, `fatal`, defaults to `info` |
| -                   | `ENABLE_METRICS`                  | `false`                    | Enable metrics collector                                     |
| -                   | `WAYBACK_LISTEN_ADDR`             | `0.0.0.0:8964`             | The listen address for the HTTP server                       |
| -                   | `CHROME_REMOTE_ADDR`              | -                          | Chrome/Chromium remote debugging address, for screenshot     |
| -                   | `WAYBACK_POOLING_SIZE`            | `3`                        | Number of worker pool for wayback at once                    |
| -                   | `WAYBACK_BOLT_PATH`               | `./wayback.db`             | File path of bolt database                                   |
| -                   | `WAYBACK_STORAGE_DIR`             | -                          | Directory to store binary file, e.g. PDF, html file          |
| -                   | `WAYBACK_MAX_MEDIA_SIZE`          | `512MB`                    | Max size to limit download stream media                      |
| -                   | `WAYBACK_MEDIA_SITES`             | -                          | Extra media websites wish to be supported, separate with comma |
| -                   | `WAYBACK_TIMEOUT`                 | `300`                      | Timeout for single wayback request, defaults to 300 second   |
| -                   | `WAYBACK_MAX_RETRIES`             | `2`                        | Max retries for single wayback request, defaults to 2        |
| -                   | `WAYBACK_USERAGENT`               | `WaybackArchiver/1.0`      | User-Agent for a wayback request                             |
| -                   | `WAYBACK_FALLBACK`                | `off`                      | Use Google cache as a fallback if the original webpage is unavailable |
| -                   | `WAYBACK_MEILI_ENDPOINT`          | -                          | Meilisearch API endpoint                                     |
| -                   | `WAYBACK_MEILI_INDEXING`          | `capsules`                 | Meilisearch indexing name                                    |
| -                   | `WAYBACK_MEILI_APIKEY`            | -                          | Meilisearch admin API key                                    |
| `-d`, `--daemon`    | -                                 | -                          | Run as daemon service, e.g. `telegram`, `web`, `mastodon`, `twitter`, `discord` |
| `--ia`              | `WAYBACK_ENABLE_IA`               | `true`                     | Wayback webpages to **Internet Archive**                     |
| `--is`              | `WAYBACK_ENABLE_IS`               | `true`                     | Wayback webpages to **Archive Today**                        |
| `--ip`              | `WAYBACK_ENABLE_IP`               | `false`                    | Wayback webpages to **IPFS**                                 |
| `--ph`              | `WAYBACK_ENABLE_PH`               | `false`                    | Wayback webpages to **[Telegra.ph](https://telegra.ph)**, required Chrome/Chromium |
| `--ipfs-host`       | `WAYBACK_IPFS_HOST`               | `127.0.0.1`                | IPFS daemon service host                                     |
| `-p`, `--ipfs-port` | `WAYBACK_IPFS_PORT`               | `5001`                     | IPFS daemon service port                                     |
| `-m`, `--ipfs-mode` | `WAYBACK_IPFS_MODE`               | `pinner`                   | IPFS mode for preserve webpage, e.g. `daemon`, `pinner`      |
| -                   | `WAYBACK_IPFS_TARGET`             | `web3storage`              | The IPFS pinning service is used to store files, supported pinners: infura, pinata, nftstorage, web3storage. |
| -                   | `WAYBACK_IPFS_APIKEY`             | -                          | Apikey of the IPFS pinning service                           |
| -                   | `WAYBACK_IPFS_SECRET`             | -                          | Secret of the IPFS pinning service                           |
| -                   | `WAYBACK_GITHUB_TOKEN`            | -                          | GitHub Personal Access Token, required the `repo` scope      |
| -                   | `WAYBACK_GITHUB_OWNER`            | -                          | GitHub account name                                          |
| -                   | `WAYBACK_GITHUB_REPO`             | -                          | GitHub repository to publish results                         |
| -                   | `WAYBACK_NOTION_TOKEN`            | -                          | Notion integration token                                     |
| -                   | `WAYBACK_NOTION_DATABASE_ID`      | -                          | Notion database ID for archiving results                     |
| `-t`, `--token`     | `WAYBACK_TELEGRAM_TOKEN`          | -                          | Telegram Bot API Token                                       |
| `--chatid`          | `WAYBACK_TELEGRAM_CHANNEL`        | -                          | The Telegram public/private channel id to publish archive result |
| -                   | `WAYBACK_TELEGRAM_HELPTEXT`       | -                          | The help text for Telegram command                           |
| -                   | `WAYBACK_MASTODON_SERVER`         | -                          | Domain of Mastodon instance                                  |
| -                   | `WAYBACK_MASTODON_KEY`            | -                          | The client key of your Mastodon application                  |
| -                   | `WAYBACK_MASTODON_SECRET`         | -                          | The client secret of your Mastodon application               |
| -                   | `WAYBACK_MASTODON_TOKEN`          | -                          | The access token of your Mastodon application                |
| -                   | `WAYBACK_TWITTER_CONSUMER_KEY`    | -                          | The customer key of your Twitter application                 |
| -                   | `WAYBACK_TWITTER_CONSUMER_SECRET` | -                          | The customer secret of your Twitter application              |
| -                   | `WAYBACK_TWITTER_ACCESS_TOKEN`    | -                          | The access token of your Twitter application                 |
| -                   | `WAYBACK_TWITTER_ACCESS_SECRET`   | -                          | The access secret of your Twitter application                |
| -                   | `WAYBACK_IRC_NICK`                | -                          | IRC nick                                                     |
| -                   | `WAYBACK_IRC_PASSWORD`            | -                          | IRC password                                                 |
| -                   | `WAYBACK_IRC_CHANNEL`             | -                          | IRC channel                                                  |
| -                   | `WAYBACK_IRC_SERVER`              | `irc.libera.chat:6697`     | IRC server, required TLS                                     |
| -                   | `WAYBACK_MATRIX_HOMESERVER`       | `https://matrix.org`       | Matrix homeserver                                            |
| -                   | `WAYBACK_MATRIX_USERID`           | -                          | Matrix unique user ID, format: `@foo:example.com`            |
| -                   | `WAYBACK_MATRIX_ROOMID`           | -                          | Matrix internal room ID, format: `!bar:example.com`          |
| -                   | `WAYBACK_MATRIX_PASSWORD`         | -                          | Matrix password                                              |
| -                   | `WAYBACK_DISCORD_BOT_TOKEN`       | -                          | Discord bot authorization token                              |
| -                   | `WAYBACK_DISCORD_CHANNEL`         | -                          | Discord channel ID, [find channel ID](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-server-ID-)  |
| -                   | `WAYBACK_DISCORD_HELPTEXT`        | -                          | The help text for Discord command                            |
| -                   | `WAYBACK_SLACK_APP_TOKEN`         | -                          | App-Level Token of Slack app                                 |
| -                   | `WAYBACK_SLACK_BOT_TOKEN`         | -                          | `Bot User OAuth Token` for Slack workspace, use `User OAuth Token` if requires create external link |
| -                   | `WAYBACK_SLACK_CHANNEL`           | -                          | Channel ID of Slack channel                                  |
| -                   | `WAYBACK_SLACK_HELPTEXT`          | -                          | The help text for Slack slash command                        |
| -                   | `WAYBACK_NOSTR_RELAY_URL`         | `wss://nostr.developer.li` | Nostr relay server url, multiple separated by comma          |
| -                   | `WAYBACK_NOSTR_PRIVATE_KEY`       | -                          | The private key of a Nostr account                           |
| `--tor`             | `WAYBACK_USE_TOR`                 | `false`                    | Snapshot webpage via Tor anonymity network                   |
| `--tor-key`         | `WAYBACK_TOR_PRIVKEY`             | -                          | The private key for Tor Hidden Service                       |
| -                   | `WAYBACK_TOR_LOCAL_PORT`          | `8964`                     | Local port for Tor Hidden Service, also support for a **reverse proxy**. This is ignored if `WAYBACK_LISTEN_ADDR` is set. |
| -                   | `WAYBACK_TOR_REMOTE_PORTS`        | `80`                       | Remote ports for Tor Hidden Service, e.g. `WAYBACK_TOR_REMOTE_PORTS=80,81` |
| -                   | `WAYBACK_TORRC`                   | `/etc/tor/torrc`           | Using `torrc` for Tor Hidden Service                         |
| -                   | `WAYBACK_SLOT`                    | -                          | Pinning service for IPFS mode of pinner, see [ipfs-pinner](https://github.com/wabarc/ipfs-pinner#supported-pinning-services) |
| -                   | `WAYBACK_APIKEY`                  | -                          | API key for pinning service                                  |
| -                   | `WAYBACK_SECRET`                  | -                          | API secret for pinning service                               |

If both of the definition file and environment variables are specified, they are all will be read and apply,
and preferred from the environment variable for the same item.

Prints the resulting options of the targets with `--print`, in a Go struct with type, without running the `wayback`.

### Docker/Podman

```sh
docker pull wabarc/wayback
docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN # without telegram channel
docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN -c YOUR-CHANNEL-USERNAME # with telegram channel
```

### 1-Click Deploy

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/wabarc/wayback)
<a href="https://render.com/deploy?repo=https://github.com/wabarc/on-render">
    <img
    src="https://render.com/images/deploy-to-render-button.svg"
    alt="Deploy to Render"
    width="155px"
    />
</a>

## Deployment

- [wabarc/on-heroku](https://github.com/wabarc/on-heroku)
- [wabarc/on-github](https://github.com/wabarc/on-github)
- [wabarc/on-render](https://github.com/wabarc/on-render)

## Documentation

For a comprehensive guide, please refer to the complete [documentation](https://docs.wabarc.eu.org/).

## Contributing

We encourage all contributions to this repository! Open an issue! Or open a Pull Request!

If you're interested in contributing to `wayback` itself, read our [contributing guide](./CONTRIBUTING.md) to get started.

Note: All interaction here should conform to the [Code of Conduct](./CODE_OF_CONDUCT.md).

## License

This software is released under the terms of the GNU General Public License v3.0. See the [LICENSE](https://github.com/wabarc/wayback/blob/main/LICENSE) file for details.

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B30014%2Fgithub.com%2Fwabarc%2Fwayback.svg?type=large)](https://app.fossa.com/projects/custom%2B30014%2Fgithub.com%2Fwabarc%2Fwayback?ref=badge_large)
