## Prerequisites

Wayback requires at least 512MB of memory, and some optional packages that can be installed below.

- [Chromium](https://www.chromium.org/Home): Wayback uses a headless Chromium to capture web pages for archiving purposes.
- [Tor](https://www.torproject.org/): Wayback can use Tor as a proxy to scrape web pages anonymously, and it can also serve as an onion service to allow users to access archived content via the Tor network.
- [youtube-dl](https://github.com/ytdl-org/youtube-dl/) or [You-Get](https://you-get.org/): Wayback can use either of these tools to download media for archiving purposes.
- [libwebp](https://developers.google.com/speed/webp/) library: Wayback uses libwebp to convert WebP images to other formats when necessary.

## Installation

The simplest, cross-platform way is to download from [GitHub Releases](https://github.com/wabarc/wayback/releases) and place the executable file in your PATH.

From source:

```sh
go install github.com/wabarc/wayback/cmd/wayback@latest
```

From GitHub Releases:

```sh
curl -fsSL https://github.com/wabarc/wayback/raw/main/install.sh | sh
```

or via [Bina](https://bina.egoist.dev/):

```sh
curl -fsSL https://bina.egoist.dev/wabarc/wayback | sh
```

Using [Snapcraft](https://snapcraft.io/wayback) (on GNU/Linux)

```sh
sudo snap install wayback
```

Via [APT](https://repo.wabarc.eu.org/deb:wayback):

```bash
curl -fsSL https://repo.wabarc.eu.org/apt/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/packages.wabarc.gpg
echo "deb [arch=amd64,arm64,armhf signed-by=/usr/share/keyrings/packages.wabarc.gpg] https://repo.wabarc.eu.org/apt/ /" | sudo tee /etc/apt/sources.list.d/wayback.list
sudo apt update
sudo apt install wayback
```

Via [RPM](https://repo.wabarc.eu.org/rpm:wayback):

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

Via [Homebrew](https://github.com/wabarc/homebrew-wayback):

```shell
brew tap wabarc/wayback
brew install wayback
```
