---
title: Web Service
---

![Web](../assets/web.png)

## 如何构建Web服务

Wayback支持服务于**明文Web**和**Onion服务**。如果缺少Tor二进制文件，则将忽略Onion服务功能。

## 配置

安装完成后，您需要通过将其放置在环境或配置文件中来提供所需的密钥。这允许您根据需要自定义配置。

- `WAYBACK_LISTEN_ADDR`：HTTP服务器的侦听地址，默认为`0.0.0.0:8964`。
- `WAYBACK_ONION_PRIVKEY`：Onion服务的私钥。
- `WAYBACK_ONION_LOCAL_PORT`：Onion服务的本地端口，也支持反向代理。如果设置了`WAYBACK_LISTEN_ADDR`，则忽略此设置。
- `WAYBACK_ONION_REMOTE_PORTS`：Onion服务的远程端口，例如`WAYBACK_ONION_REMOTE_PORTS=80,81`。

注意：要首次运行Onion服务，您需要保留`私钥`，该私钥可以从日志输出中看到。
