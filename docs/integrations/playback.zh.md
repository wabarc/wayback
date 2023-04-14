---
title: Playback in wayback
---

Wayback现在支持从多个归档服务中进行回放。对于[IPFS](./ipfs.md)和[Telegraph](./telegraph.md)，您需要设置自托管服务，可以是Meilisearch或GitHub Issues。

目前，以下归档服务支持在wayback中进行回放：

- [Google Cache](https://webcache.googleusercontent.com/)
- [Internet Archive](https://web.archive.org/)
- [IPFS](https://ipfs.github.io/public-gateway-checker/)
- [archive.today](https://archive.today/)
- [Telegraph](https://telegra.ph/)
- [Time Travel](http://timetravel.mementoweb.org/)

## 如何配置回放服务

将这些密钥放置在环境或配置文件中：

对于Meilisearch（推荐）：

- `PLAYBACK_MEILI_ENDPOINT`：Meilisearch API端点。
- `PLAYBACK_MEILI_INDEXING`：Meilisearch索引名称，默认为`capsules`（可选）。
- `PLAYBACK_MEILI_APIKEY`：Meilisearch管理员API密钥（可选）。

对于GitHub Issues：

- `PLAYBACK_GITHUB_REPO`：GitHub存储库以发布结果。
- `PLAYBACK_GITHUB_PAT`：GitHub个人访问令牌，用于降低GitHub API请求的速率限制。它需要`repo`作用域。

## 回放集成

回放无缝集成到Wayback中，以增强用户体验。它目前支持从各种平台进行回放，包括Discord、Mastodon、Matrix、Web、Slack和Telegram。

### Web

要回放归档内容，请访问[Clear Web](https://wabarc.eu.org/)或[Onion Service](http://wabarcoww2bxmdbixj7sjwggv3fonh2rpflfiildegcydk5udkdckdyd.onion/)。
输入要回放的URI。您可以输入多个URI，用逗号或换行符分隔。
单击左下角的按钮，等待几秒钟，然后开始回放。

### 即时通讯工具

即时通讯工具，如Discord、Matrix、Slack和Telegram，都具有回放功能，可以使用相同的命令访问：`/playback`。
输入`/playback https://example.com`，然后等待响应。

### 其他

要在Mastodon上使用Playback，请向机器人发送带有关键字`/playback`和要回放的URI的消息。
然后，等一会儿，机器人就会回复。
