---
title: 归档到IPFS
---

Wayback依赖于星际文件系统([IPFS](https://ipfs.tech/))作为存储完整网页及相关资源（如JavaScript、CSS和字体）的上游服务。这样可以无缝地播放存档的网页，确保用户体验与原始网站完全相同。

您可以使用`--ip`标志或默认启用的`WAYBACK_ENABLE_IP`环境变量启用或禁用此功能。

Wayback实现IPFS集成的代码可以在[wabarc/rivet](https://github.com/wabarc/rivet)存储库中找到。
