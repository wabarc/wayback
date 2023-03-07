---
title: Publish to Nostr
---

## How to build a Nostr Bot

Wayback currently only supports publishing to Nostr as the Nostr Protocol is still under development.

Select any relay to generate a private key (here's a [guide](https://nostr.how/) to help you get started).

## Configuration

After creating a new account, you will have the `private key`.

Next, place these keys in the environment or configuration file:

- `WAYBACK_NOSTR_RELAY_URL`: Nostr relay server url, multiple separated by comma.
- `WAYBACK_NOSTR_PRIVATE_KEY`: The private key of a Nostr account.

## Further reading

- [Nostr Protocol](https://github.com/nostr-protocol/nostr)
