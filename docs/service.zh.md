## 配置

Wayback提供了两种配置格式：**环境变量**和**配置文件**，它们使用相同的键。配置文件采用INI文件格式，并可在[wayback.conf](https://github.com/wabarc/wayback/blob/main/wayback.conf)中找到。环境变量可以使用前缀`WAYBACK_`后跟配置键名来设置。

例如，要同时提供Discord和Telegram机器人服务，您可以运行命令`wayback -d discord -d telegram`。`-d`标志后跟平台的名称，应该是小写的。

要查看有关可用CLI标志的其他信息，请运行命令`wayback --help`。

## 服务

Wayback可以与各种消息平台集成，包括Discord、IRC、Mastodon、Matrix、Slack、Telegram、Twitter和Web，作为响应用户查询的机器人。

有关如何为每个平台创建机器人的详细说明，请参见以下链接：

- [Discord](integrations/discord.md)
- [IRC](integrations/irc.md)
- [Mastodon](integrations/mastodon.md)
- [Matrix](integrations/matrix.md)
- [Slack](integrations/slack.md)
- [Telegram](integrations/telegram.md)
- [Twitter](integrations/twitter.md)
- [Web](integrations/web.md)
- [XMPP](integrations/xmpp.md)

请注意，您需要在各自的平台上设置帐户并获取必要的凭据，例如访问令牌，才能将Wayback用作机器人。

## 发布

Wayback的集成服务提供了将存档结果发布到各种消息和协作平台的功能。发布的结果不包括任何请求者信息，以确保隐私。

**当您放置必要的配置时，该项的发布功能将自动启用。**

有关如何配置发布通道的详细说明，请参见以下链接：

- [IRC](integrations/irc.md)
- [Discord](integrations/discord.md)
- [GitHub Issues](integrations/github.md)
- [Mastodon](integrations/mastodon.md)
- [Matrix](integrations/matrix.md)
- [Meilisearch](integrations/meilisearch.md)
- [Nostr](integrations/nostr.md)
- [Notion](integrations/notion.md)
- [Omnivore](integrations/omnivore.md)
- [Postgres](integrations/datastore.md)
- [Slack](integrations/slack.md)
- [Telegram](integrations/telegram.md)
- [Twitter](integrations/twitter.md)

每个平台都有自己的配置要求，因此请务必仔细按照说明操作，以确保成功发布存档结果。
