---
title: Deployment
---

Wayback can be deployed using Docker/Podman containers, and can also be easily deployed on cloud platforms such as Heroku and Render.

## Docker/Podman

```sh
docker pull wabarc/wayback
docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN # without telegram channel
docker run -d wabarc/wayback wayback -d telegram -t YOUR-BOT-TOKEN -c YOUR-CHANNEL-USERNAME # with telegram channel
```

## 1-Click Deploy

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/wabarc/wayback)
<a href="https://render.com/deploy?repo=https://github.com/wabarc/on-render">
    <img
    src="https://render.com/images/deploy-to-render-button.svg"
    alt="Deploy to Render"
    width="155px"
    />
</a>

## Deployment

- [wabarc/on-heroku](https://github.com/wabarc/on-heroku)
- [wabarc/on-github](https://github.com/wabarc/on-github)
- [wabarc/on-render](https://github.com/wabarc/on-render)
