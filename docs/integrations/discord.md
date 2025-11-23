---
title: Interactive with Discord
---

![Discord](../assets/discord-server.png)

## How to build a Discord Bot

To build a Discord bot, you will need to follow these steps:

Create a [Discord application](https://discord.com/developers/applications) with the `bot` and `applications.commands` scopes enabled in the `OAuth2 - SCOPES` section. Make sure to grant the bot the permissions to `Send Messages` and `Attach Files`.

Configure the Discord bot to support the following slash commands:

1. `/help` - shows help information (*configured help text is required*)
2. `/metrics` - shows service metrics (*enabled metrics is required*)
3. `/playback` - playback URLs

Set up the following environment variables for configuring a Discord daemon service:

1. `WAYBACK_DISCORD_TOKEN` (required)
2. `WAYBACK_DISCORD_CHANNEL`
3. `WAYBACK_DISCORD_HELPTEXT`

For detailed documentation on how to create and configure a Discord bot, please see the [Discord Developer Portal](https://discord.com/developers/docs/intro).
