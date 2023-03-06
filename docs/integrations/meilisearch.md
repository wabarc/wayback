---
title: Publish to Meilisearch
---

## How to build a service

Meilisearch is a fast and powerful open-source search engine that can deliver relevant search results in a matter of milliseconds. Since [v0.18.0](https://github.com/wabarc/wayback/releases/tag/v0.18.0), wayback has been supports Meilisearch to store archived results for playback. The following data structure is used:

```proto
message Document {
    string Source = 1;
    string IA = 2;
    string IS = 3;
    string IP = 4;
    string PH = 5;
}
```

To install Meilisearch, you can follow the installation guide available on the official Meilisearch website: <https://docs.meilisearch.com/learn/getting_started/installation.html>.

## Configuration

After running Meilisearch, you will have the endpoint.

Next, place these keys in the environment or configuration file:

- `WAYBACK_MEILI_ENDPOINT`: Meilisearch API endpoint.
- `WAYBACK_MEILI_INDEXING`: Meilisearch indexing name, defaults to `capsules` (optional).
- `WAYBACK_MEILI_APIKEY`: Meilisearch admin API key (optional).
