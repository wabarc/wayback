---
title: 发布到数据库
---

注意：仅支持 Postgres。

## 配置

- `WAYBACK_DATABASE_URL`： Postgres 数据库的 URL，例如 `user=postgres password=postgres dbname=wayback sslmode=disable`。
- `WAYBACK_DATABASE_MAX_CONNS`： Postgres 数据库的最大连接数（可选）。
- `WAYBACK_DATABASE_MIN_CONNS`： Postgres 数据库的最小连接数（可选）。
- `WAYBACK_DATABASE_CONNECTION_LIFETIME`： Postgres 数据库的连接生命周期（可选）。

