---
title: Playback in wayback
---

Wayback now supports playback from multiple archiving services. For [IPFS](./ipfs.md) and [Telegraph](./telegraph.md), you will need to set up self-hosted services, which can be Meilisearch or GitHub Issues.

The following archiving services are currently supported for playback in wayback:

- [Google Cache](https://webcache.googleusercontent.com/)
- [Internet Archive](https://web.archive.org/)
- [IPFS](https://ipfs.github.io/public-gateway-checker/)
- [archive.today](https://archive.today/)
- [Telegraph](https://telegra.ph/)
- [Time Travel](http://timetravel.mementoweb.org/)

## How to config a playback service

Place these keys in the environment or configuration file:

For Meilisearch (recommended):

- `PLAYBACK_MEILI_ENDPOINT`: Meilisearch API endpoint.
- `PLAYBACK_MEILI_INDEXING`: Meilisearch indexing name, defaults to `capsules` (optional).
- `PLAYBACK_MEILI_APIKEY`: Meilisearch admin API key (optional).

For GitHub Issues:

- `PLAYBACK_GITHUB_REPO`: GitHub repository to publish results.
- `PLAYBACK_GITHUB_PAT`: GitHub Personal Access Token, which is used to reduce the rate limit for GitHub API requests. It requires the `repo` scope.

## Playback Integrations

Playback is seamlessly integrated into Wayback to enhance the user experience. It currently supports playback from various platforms, including Discord, Mastodon, Matrix, Web, Slack, and Telegram.

### Web

1. To playback archived content, visit either the [Clear Web](https://wabarc.eu.org/) or [Onion Service](http://wabarcoww2bxmdbixj7sjwggv3fonh2rpflfiildegcydk5udkdckdyd.onion/).
2. Enter the URIs you want to play back. You can enter multiple URIs separated by commas or line breaks.
3. Click the button in the bottom left corner and wait a few moments for the playback to begin.

### Instant messaging tools

Instant messaging tools such as Discord, Matrix, Slack and Telegram all have a playback feature that can be accessed using the same command: `/playback`.
Enter `/playback https://example.com` and wait for the response.

### Others

To use Playback on Mastodon, send a message to the bot with the keyword `/playback` at the beginning and the URI you wish to playback.
Then, wait a moment for the bot to respond.
