# Configuration Parameters

Wayback can use a configuration file and environment variables.

If both of the definition file and environment variables are specified, they are all will be read and apply,
and preferred from the environment variable for the same item.

Prints the resulting options of the targets with `--print`, in a Go struct with type, without running the `wayback`.

By default, `wayback` looks for configuration options from this files, the following are parsed:

- `./wayback.conf`
- `~/wayback.conf`
- `/etc/wayback.conf`

Use the `-c` / `--config` option to specify the build definition file to use.

## Configuration Options

| Flags               | Environment Variable              | Default                    | Description                                                  |
| ------------------- | --------------------------------- | -------------------------- | ------------------------------------------------------------ |
| `--debug`           | `DEBUG`                           | `false`                    | Enable debug mode, override `LOG_LEVEL`                      |
| `-c`, `--config`    | -                                 | -                          | Configuration file path, defaults: `./wayback.conf`, `~/wayback.conf`, `/etc/wayback.conf` |
| -                   | `LOG_TIME`                        | `true`                     | Display the date and time in log messages                    |
| -                   | `LOG_LEVEL`                       | `info`                     | Log level, supported level are `debug`, `info`, `warn`, `error`, `fatal`, defaults to `info` |
| -                   | `ENABLE_METRICS`                  | `false`                    | Enable metrics collector                                     |
| -                   | `WAYBACK_LISTEN_ADDR`             | `0.0.0.0:8964`             | The listen address for the HTTP server                       |
| -                   | `CHROME_BIN`                      | -                          | Preferred to sets the path to the Chrome executable          |
| -                   | `CHROME_REMOTE_ADDR`              | -                          | Chrome/Chromium remote debugging address, for screenshot, format: `host:port`, `wss://domain.tld` |
| -                   | `WAYBACK_PROXY`                   | -                          | Proxy address, e.g. `socks5://127.0.0.1:1080`                |
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
| -                   | `WAYBACK_XMPP_JID`                | -                          | The JID of a XMPP account                                    |
| -                   | `WAYBACK_XMPP_PASSWORD`           | -                          | The password of a XMPP account                               |
| -                   | `WAYBACK_XMPP_NOTLS`              | -                          | Connect to XMPP server without TLS                           |
| -                   | `WAYBACK_XMPP_HELPTEXT`           | -                          | The help text for XMPP command                               |
| `--tor`             | `WAYBACK_USE_TOR`                 | `false`                    | Snapshot webpage via Tor anonymity network                   |
| `--tor-key`         | `WAYBACK_ONION_PRIVKEY`           | -                          | The private key for Tor Hidden Service                       |
| -                   | `WAYBACK_ONION_LOCAL_PORT`        | `8964`                     | Local port for Tor Hidden Service, also support for a **reverse proxy**. This is ignored if `WAYBACK_LISTEN_ADDR` is set. |
| -                   | `WAYBACK_ONION_REMOTE_PORTS`      | `80`                       | Remote ports for Tor Hidden Service, e.g. `WAYBACK_ONION_REMOTE_PORTS=80,81` |
| -                   | `WAYBACK_ONION_DISABLED`          | `false`                    | Disable onion service                                        |
| -                   | `WAYBACK_SLOT`                    | -                          | Pinning service for IPFS mode of pinner, see [ipfs-pinner](https://github.com/wabarc/ipfs-pinner#supported-pinning-services) |
| -                   | `WAYBACK_APIKEY`                  | -                          | API key for pinning service                                  |
| -                   | `WAYBACK_SECRET`                  | -                          | API secret for pinning service                               |
