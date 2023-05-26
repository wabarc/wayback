---
title: 简介
---

Wayback是一个网络归档和回放工具，允许用户抓取和保存网络内容。它提供了一个IM风格的界面来接收和展示存档的网络内容，以及一个搜索和回放服务来检索以前存档的网页。Wayback是为网络档案员、研究人员和任何想保存网络内容并在未来访问这些内容的人设计的。

Wayback是用Go编写的开源网络存档应用程序。具有模块化和可定制化的架构，它被设计成灵活和适应各种用例和环境。它提供了对多个存储后端的支持，并与其他服务集成。

无论您是需要归档一个网页，还是需要归档一大批网站，Wayback都可以帮助你抓取和保存网络内容，以备后用。凭借易于使用的界面和强大的功能，Wayback对于任何对网络归档和保存感兴趣的人来说都是一个宝贵的工具。

## 特性

- 完全开源
- 跨平台兼容
- 输出Prometheus度量指标
- 批量存档URL以加快存档速度
- 内置CLI工具（`wayback`）以便于使用
- 可作为Tor隐藏服务或本地Web入口，增加隐私和可访问性
- 更容易地集成到Internet Archive、archive.today、IPFS和Telegraph中
- 可与IRC、Matrix、Telegram机器人、Discord机器人、Mastodon、Twitter和XMPP进行交互，作为守护进程服务，以便于使用
- 支持将存档结果发布到Telegram频道、Mastodon和GitHub Issues中进行共享
- 支持将存档文件存储到磁盘中以供离线使用
- 下载流媒体（需要[FFmpeg](https://ffmpeg.org/)）以便于媒体存档。

## 工作原理

![How wayback works](./assets/wayback.svg "How wayback works")
