---
title: 发布到Meilisearch
---

## 如何构建服务

Meilisearch是一款快速而强大的开源搜索引擎，可以在毫秒级别内提供相关的搜索结果。从[v0.18.0](https://github.com/wabarc/wayback/releases/tag/v0.18.0)版本开始，wayback支持使用Meilisearch存储归档结果以进行回放。使用以下数据结构：

```proto
message Document {
    string Source = 1;
    string IA = 2;
    string IS = 3;
    string IP = 4;
    string PH = 5;
}
```

要安装Meilisearch，您可以按照官方Meilisearch网站上的安装指南进行操作：<https://docs.meilisearch.com/learn/getting_started/installation.html>。

## 配置

运行Meilisearch后，您将拥有终端节点。

接下来，将这些密钥放置在环境或配置文件中：

- `WAYBACK_MEILI_ENDPOINT`：Meilisearch API终端节点。
- `WAYBACK_MEILI_INDEXING`：Meilisearch索引名称，默认为`capsules`（可选）。
- `WAYBACK_MEILI_APIKEY`：Meilisearch管理员API密钥（可选）。
