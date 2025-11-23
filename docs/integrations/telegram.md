---
title: Interactive with Telegram
---

![Telegram Bot](../assets/telegram.png)

## How to build a Telegram Bot

Steps to create a new bot:

1. Open the Telegram app and search for the BotFather bot.
2. Start a chat with BotFather by clicking the "Start" button.
3. Send the command `/newbot` to BotFather to create a new bot.
4. Follow the instructions from BotFather and provide a name and username for your bot.
5. After creating the bot, BotFather will provide you with a token.
6. To test your bot, open a chat with your bot by searching for its username and send a message.
7. You can also customize your bot by sending commands to BotFather, such as `/setdescription` and `/setuserpic`.

Optionally, you can also create a channel for publishing.

## Configuration

After creating a new bot, you will have the `Bot API Token`.

Next, place these keys in the environment or configuration file:

- `WAYBACK_TELEGRAM_TOKEN`: Bot API Token.
- `WAYBACK_TELEGRAM_CHANNEL`: Channel ID for publishing (optional).
- `WAYBACK_TELEGRAM_HELPTEXT`: Provide a help message for users to reference (optional).

## Further reading

- [Bots: An introduction for developers](https://core.telegram.org/bots)
