## 前提条件

Wayback需要至少512MB的内存，并且可以安装以下一些可选软件包。

- [Chromium](https://www.chromium.org/Home): Wayback使用一个无头Chromium来捕获网页以进行存档。
- [Tor](https://www.torproject.org/): Wayback可以使用Tor作为代理以匿名地爬取网页，同时也可以作为一个洋葱服务，允许用户通过Tor网络访问存档内容。
- [youtube-dl](https://github.com/ytdl-org/youtube-dl/) or [You-Get](https://you-get.org/): Wayback可以使用这些工具之一来下载媒体以进行存档。
- [libwebp](https://developers.google.com/speed/webp/) library: Wayback在必要时使用libwebp将WebP图像转换为其他格式。

## 安装

最简单的跨平台方式是从[GitHub发布页面](https://github.com/wabarc/wayback/releases)下载并将可执行文件放置于您的**PATH**变量中。

从源码安装:

```sh
go install github.com/wabarc/wayback/cmd/wayback@latest
```

下载预先编译的二进制文件:

```sh
curl -fsSL https://github.com/wabarc/wayback/raw/main/install.sh | sh
```

通过 [Bina](https://bina.egoist.dev/):

```sh
curl -fsSL https://bina.egoist.dev/wabarc/wayback | sh
```

使用 [Snapcraft](https://snapcraft.io/wayback) (on GNU/Linux)

```sh
sudo snap install wayback
```

通过 [APT](https://repo.wabarc.eu.org/deb:wayback):

```bash
curl -fsSL https://repo.wabarc.eu.org/apt/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/packages.wabarc.gpg
echo "deb [arch=amd64,arm64,armhf signed-by=/usr/share/keyrings/packages.wabarc.gpg] https://repo.wabarc.eu.org/apt/ /" | sudo tee /etc/apt/sources.list.d/wayback.list
sudo apt update
sudo apt install wayback
```

通过 [RPM](https://repo.wabarc.eu.org/rpm:wayback):

```bash
sudo rpm --import https://repo.wabarc.eu.org/yum/gpg.key
sudo tee /etc/yum.repos.d/wayback.repo > /dev/null <<EOT
[wayback]
name=Wayback Archiver
baseurl=https://repo.wabarc.eu.org/yum/
enabled=1
gpgcheck=1
gpgkey=https://repo.wabarc.eu.org/yum/gpg.key
EOT

sudo dnf install -y wayback
```

通过 [Homebrew](https://github.com/wabarc/homebrew-wayback):

```shell
brew tap wabarc/wayback
brew install wayback
```
