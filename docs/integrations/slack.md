---
title: Interactive with Slack
---

## How to build a Slack App

Steps to create a new app:

1. Open [Slack API](https://api.slack.com/apps).
2. Click "Create New App" and "From Scratch".
2. Generate an App-Level Token with the `connections:write` scope.
3. Enable Socket Mode.
4. Enable Events
    - Subscribe to bot events: `app_mention` and `message.im`.
    - Subscribe to events on behalf of users: `message.im`.
5. Setting OAuth & Permissions User Token Scopes: `chat:write`, `files:write`.
6. Install the app to your workspace and obtain the `Bot User OAuth Token`.
7. In the App Home section, check `Allow users to send Slash commands and messages from the messages tab`.
8. Optionally, create a channel for publishing and note down the `Channel ID` by viewing the channel details.

## Configuration

After creating a new app, you will have the `Bot User OAuth Token` and `Channel ID`.

Next, place these keys in the environment or configuration file:

- `WAYBACK_SLACK_BOT_TOKEN`: Bot User OAuth Token.
- `WAYBACK_SLACK_CHANNEL`: Channel ID for publishing (optional).
- `WAYBACK_SLACK_HELPTEXT`: Provide a help message for users to reference (optional).

## Further reading

- [Slack API Documentation](https://api.slack.com/)
