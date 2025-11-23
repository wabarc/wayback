---
title: Telegram
---

![Telegram Bot](../assets/telegram.png)

## 如何构建一个Telegram机器人

创建新机器人的步骤如下：

1. 打开Telegram应用并搜索BotFather机器人。
2. 点击“开始”按钮与BotFather开始聊天。
3. 向BotFather发送命令`/newbot`来创建一个新机器人。
4. 按照BotFather的指示提供机器人的名称和用户名。
5. 创建机器人后，BotFather将为您提供一个令牌。
6. 要测试您的机器人，请通过搜索其用户名打开与您的机器人的聊天，并发送一条消息。
7. 您还可以通过向BotFather发送命令，例如`/setdescription`和`/setuserpic`，来定制您的机器人。

可选地，您还可以创建一个用于发布的频道。

## 配置

创建新机器人后，您将获得`Bot API Token`。

接下来，将这些密钥放置在环境或配置文件中：

- `WAYBACK_TELEGRAM_TOKEN`：Bot API Token。
- `WAYBACK_TELEGRAM_CHANNEL`：用于发布的频道ID（可选）。
- `WAYBACK_TELEGRAM_HELPTEXT`：提供帮助消息供用户参考（可选）。

## 相关资料

- [开发人员入门：Bots](https://core.telegram.org/bots)
