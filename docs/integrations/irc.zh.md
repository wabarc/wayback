---
title: IRC
---

## 如何构建IRC机器人

要创建一个IRC帐户，您可以访问Libera.Chat的[昵称注册](https://libera.chat/guides/registration)页面以获取说明。

## 配置

为了使用IRC服务，您需要设置以下环境变量或配置文件：

- `WAYBACK_IRC_NICK`：IRC机器人的昵称（必填）。
- `WAYBACK_IRC_PASSWORD`：IRC机器人的密码（可选）。
- `WAYBACK_IRC_CHANNEL`：发布存档结果的频道（可选）。
- `WAYBACK_IRC_SERVER`：连接到的IRC服务器（可选，默认为`irc.libera.chat:6697`）。

请注意，一些IRC服务器可能需要额外的配置，例如TLS证书。您应该参考您想要连接的IRC服务器的文档以获取更多信息。

## 相关资料
- [IRC命令](https://en.wikipedia.org/wiki/List_of_Internet_Relay_Chat_commands)
- [IRCv3规范](https://ircv3.net/irc/)
- [IRC网络列表](https://netsplit.de/networks/top100.php)
- [Libera.chat文档](https://libera.chat/guides)
