---
title: Publish to Database
---

Note: Only Postgres is supported.

## Configuration

- `WAYBACK_DATABASE_URL`: The URL of the Postgres database, e.g. `user=postgres password=postgres dbname=wayback sslmode=disable`.
- `WAYBACK_DATABASE_MAX_CONNS`: Maximum connections of the Postgres database (optional).
- `WAYBACK_DATABASE_MIN_CONNS`: Minimum connections of the Postgres database (optional).
- `WAYBACK_DATABASE_CONNECTION_LIFETIME`: Connection lifetime of the Postgres database (optional).

