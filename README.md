# Wayback

`wabarc/wayback` is a tool that supports running as a command-line tool and docker container, purpose to snapshot webpage to time capsules.

## Prerequisites

- Golang
- Telegram bot

## Installation

```sh
$ go get -u github.com/wabarc/wayback
```

## Usage

1. Running the command-line or Docker container.
2. Start a chat with the bot and Send URL.

### Command line

```sh
$ wayback -h
A CLI tool for wayback webpages.

Usage:
  wayback [flags]
  wayback [command]

Available Commands:
  help        Help about any command
  telegram    A CLI tool for wayback webpages on Telegram bot.

Flags:
  -h, --help   help for wayback

Use "wayback [command] --help" for more information about a command.

$ wayback telegram -t YOUR-BOT-TOKEN
```

Also can run with debug mode:

```sh
$ wayback telegram -t YOUR-BOT-TOKEN -d
```

### Docker/Podman

```sh
$ docker pull wabarc/wayback
$ docker run -d wabarc/wayback telegram -t YOUR-BOT-TOKEN
```

## TODO

[Archive.org](https://web.archive.org/) and [Archive.today](https://archive.today/) are currently supported, the next step mind support the followings platform:

- [IPFS](https://ipfs.io/)
- [ZeroNet](https://zeronet.io/)

## Telegram bot

- [Bots: An introduction for developers](https://core.telegram.org/bots)
- [How do I create a bot?](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- [An example bot](http://t.me/wabarc_bot)

## License

Permissive GPL 3.0 license, see the [LICENSE](https://github.com/wabarc/wayback/blob/master/LICENSE) file for details.
