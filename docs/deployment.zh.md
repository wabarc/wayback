---
title: 部署
---

Wayback可以使用Docker/Podman容器部署，也可以轻松部署在Heroku和Render等云平台上。

## Docker/Podman

```sh
docker pull wabarc/wayback
docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN # without telegram channel
docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN -c YOUR-CHANNEL-USERNAME # with telegram channel
```

## 一键部署

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/wabarc/wayback)
<a href="https://render.com/deploy?repo=https://github.com/wabarc/on-render">
    <img
    src="https://render.com/images/deploy-to-render-button.svg"
    alt="Deploy to Render"
    width="155px"
    />
</a>

## 快速部署

- [wabarc/on-heroku](https://github.com/wabarc/on-heroku)
- [wabarc/on-github](https://github.com/wabarc/on-github)
- [wabarc/on-render](https://github.com/wabarc/on-render)
