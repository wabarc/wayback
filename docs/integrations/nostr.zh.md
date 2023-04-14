---
title: 发布到Nostr
---

## 如何构建Nostr机器人

目前，Wayback仅支持将归档结果发布到Nostr，因为Nostr协议仍在开发中。

选择任何继电器以生成一个私钥（这里有一个[指南](https://nostr.how/)可以帮助您入门）。

## 配置

创建新帐户后，您将拥有`私钥`。

接下来，将这些密钥放置在环境或配置文件中：

- `WAYBACK_NOSTR_RELAY_URL`：Nostr中继服务器URL，用逗号分隔的多个URL。
- `WAYBACK_NOSTR_PRIVATE_KEY`：Nostr帐户的私钥。

## 相关资料

- [Nostr协议](https://github.com/nostr-protocol/nostr)
