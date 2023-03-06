## Configuration

Wayback provides two configuration formats: **environment variables** and a **configuration file**, which use the same keys. The configuration file is in the INI file format and can be found at [wayback.conf](https://github.com/wabarc/wayback/blob/main/wayback.conf). The environment variables can be set with the prefix `WAYBACK_` followed by the configuration key name.

For example, to serve both a Discord and a Telegram bot, you can run the command `wayback -d discord -d telegram`. The `-d` flag is followed by the name of the platform, which should be in lowercase.

To view additional information about the available CLI flags, run the command `wayback --help`.

## Service

Wayback can be integrated with various messaging platforms, including Discord, IRC, Mastodon, Matrix, Slack, Telegram, Twitter and Web, to function as a bot that responds to user queries.

For detailed instructions on how to create a bot for each platform, please refer to the links below:

- [Discord](integrations/discord.md)
- [IRC](integrations/irc.md)
- [Mastodon](integrations/mastodon.md)
- [Matrix](integrations/matrix.md)
- [Slack](integrations/slack.md)
- [Telegram](integrations/telegram.md)
- [Twitter](integrations/twitter.md)
- [Web](integrations/web.md)
- [XMPP](integrations/xmpp.md)

Please note that you need to set up accounts on the respective platforms and obtain necessary credentials, such as access tokens, to use Wayback as a bot.

## Publish

Wayback's integrated services provide the ability to publish archiving results to various messaging and collaboration platforms. The published results do not include any requester information to ensure privacy.

**When you place the necessary configuration, the publish feature for that item will be automatically enabled.**

For detailed instructions on how to configure the publishing channel, please refer to the links below:

- [IRC](integrations/irc.md)
- [Discord](integrations/discord.md)
- [GitHub Issues](integrations/github.md)
- [Mastodon](integrations/mastodon.md)
- [Matrix](integrations/matrix.md)
- [Meilisearch](integrations/meilisearch.md)
- [Nostr](integrations/nostr.md)
- [Notion](integrations/notion.md)
- [Slack](integrations/slack.md)
- [Telegram](integrations/telegram.md)
- [Twitter](integrations/twitter.md)

Each platform has its own configuration requirements, so be sure to follow the instructions carefully to ensure successful publishing of archiving results.
