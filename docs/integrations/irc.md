---
title: Interactive with IRC
---

## How to build a IRC Bot

To create an IRC account, you can visit the [Nickname Registration](https://libera.chat/guides/registration) page on Libera.Chat for instructions.

## Configuration

To use the IRC service, you will need to set the following environment variables or configuration file:

- `WAYBACK_IRC_NICK`: The nickname for the IRC bot (required).
- `WAYBACK_IRC_PASSWORD`: The password for the IRC bot (optional).
- `WAYBACK_IRC_CHANNEL`: The channel for publish the archiving results (optional).
- `WAYBACK_IRC_SERVER`: The IRC server to connect to (optional, defaults to `irc.libera.chat:6697`).

Note that some IRC servers may require additional configuration, such as TLS certificates. You should refer to the documentation of the IRC server you wish to connect to for more information.

## Further reading
- [IRC commands](https://en.wikipedia.org/wiki/List_of_Internet_Relay_Chat_commands)
- [IRCv3 Specifications](https://ircv3.net/irc/)
- [List of IRC networks](https://netsplit.de/networks/top100.php)
- [Libera.chat documentation](https://libera.chat/guides)
