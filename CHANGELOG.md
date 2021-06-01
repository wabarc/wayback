# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.12.3] - 2021-06-01

### Added
- Add Dockerfile for development

### Changed
- Refine artifact name for testing workflow
- Supports specify boltdb file path
- Improve metrics of web entry

### Fixed
- Fix multiple results for archiving pdf file ([wabarc/screenshot](https://github.com/wabarc/screenshot/commit/f17a852a3ae2a7c9396719c526f7bd8f2688bbe2))

## [0.12.2] - 2021-05-26

### Changed
- Stability improvements on [wabarc/telegra.ph](https://github.com/wabarc/telegra.ph/commit/85ca843f66376b2ebcd2235762dab75694b6a3e6)
- Reply queue message from Telegram
- Upgrade linter to v4
- Update README

## [0.12.1] - 2021-05-23

### Changed
- Improvement for illegal command
- Enhancement for Tor Hidden Service
- Set defaults IRC server to Libera Chat
- Styling code base

### Fixed
- Prevent dispatch multiple deployment
- Fix release notes announcements

## [0.12.0] - 2021-05-19

### Added
- Add worker pool
- Handle message from Telegram group
- Add APT, RPM and Homebrew repository
- Publish release note to Telegram channel

### Changed
- Handle mastodon message using notification instead conversation
- Packaging license, changelog and readme
- Improve web layout

## [0.11.1] - 2021-05-12

### Added
- Store playback data locally
- Auto fallback to Google cache if URI is missing

### Changed
- Migrate telegram-bot-api to telebot, support auto append bot command
- Update PAT to GITHUB_TOKEN
- Exclude path from service worker
- Upgrade dependencies
- Minor improvements

## [0.11.0] - 2021-05-06

### Added
- Add PWA support
- Add more tests
- Build package for Archlinux
- Setup tor for testing workflow
- Generate Git log as release note
- Dispatch repository in wabarc/on-heroku
- Add Heroku process file
- Add metrics collector

### Changed
- Join IRC channel before connect
- Doesn't reply if a forwarded message from telegram without caption
- Attach a button below the message for send a wayback request
- Upgrade dependencies
- Refactor archive func
- Close services using context cancellation signals
- Check defaults port idle status to use torrc
- Append defaults telegram command to fallback text

### Fixed
- Validate text for publish
- Fix template render without args

### Removed
- Remove defaults command `/search` and `/status` for telegram

## [0.10.3] - 2021-04-21

### Changed
- Validate URL for render message.
- Improve playback for telegram.
- Use Google document viewer to open files. </wabarc/screenshot/releases/tag/v1.1.1>

## [0.10.2] - 2021-04-20

### Added
- Support screenshot using Chrome remote debugging address.

### Changed
- Improve telegram command message.
- Append title content from `og:title` if empty.
- Use socks proxy for `archive.is` as defaults.

## [0.10.1] - 2021-04-18

### Changed
- Update Dockerfile label
- Update Telegram message template

### Fixed
- Fix publish in multiple mode

## [0.10.0] - 2021-04-17

### Added
- Add flag `-c` and `--config` to specify configuration file path.
- Add tests for publish.
- Add playback for Telegram bot.
- Supports to set help command for Telegram bot.

### Changed
- Refactor configuration handler.
- Redact message without URL for Matrix.
- Separate logger package.

### Removed
- Remove flag `-c` to define Telegram channel name.

### Fixed
- Fix Matrix RoomID format.

## [0.9.1] - 2021-04-12

### Fixed
- Fix publish context panic.

## [0.9.0] - 2021-04-12

### Added
- Add IRC support.
- Add Matrix support.
- Add linter rules for workflow.
- Add reviewdog workflow.
- Build binary for Apple Silicon.
- Build binary for FreeBSD/arm64.

### Changed
- Refine Dockerfile.
- Refine test workflow.
- Improve Docker image release workflow.
- Upgrade dependencies.
- Listen on local port `8964` for web service.

## [0.8.3] - 2021-03-24

### Added
- Add test for twitter service.
- Build multi-arch deb package.

### Changed
- Refactor publish service.

### Fixed
- Minor bugfix.

## [0.8.2] - 2021-03-05

### Changed
- Update man page.
- Add more exclude exit nodes of Tor for Docker image.
- Styling output results in command.
- Set Tor temporary data directory.

## [0.8.1] - 2021-03-02

### Added
- Clear Mastodon notifications every 10 minutes.
- Handle os signal.

### Changed
- Adjust request Mastodon API interval to 5 seconds.
- Upgrade RPM builder Go version to 1.16

### Fixed
- Fix nil pointer dereference of archive.today.

## [0.8.0] - 2021-02-27

### Added
- Add Mastodon support.
- Supports publish toot even if the entry from Telegram Bot and Tor Hidden Service.
- Add Twitter support.
- Supports publish tweet even if the entry from Mastodon Bot, Telegram Bot and Tor Hidden Service.
- Add stale workflow.

### Changed
- Make logs more readable.
- Update snapcraft workflow.

## [0.7.0] - 2021-02-24

### Added
- Add publish to GitHub Issues support.

### Changed
- Styling channel message.

## [0.6.3] - 2021-02-21

### Changed
- Upload image to ImgBB.
- Set image quality to 100.
- Upgrade Go version to 1.16

### Fixed
- Fix create telegra.ph page failure due to title too long.

## [0.6.0] - 2021-01-28

### Added
- Add wayback to Telegraph support.

### Changed
- Using `/etc/tor/torrc` for Tor Hidden Service via the `WAYBACK_TORRC` environment variable

### Fixed
- Minor bugfixs.

## [0.5.6] - 2021-01-24

### Changed
- Now available to access the archive.today's tor service if enable service of archive.today.
- Publish multiple arch snapcraft app.
- Refine workflows.

### Fixed
- Fix telegram user id conflict in reply.
- Fix nil pointer dereference.

## [0.5.5] - 2021-01-15

### Added
- Support publish message to channel with Tor entry.

### Fixed
- Minor bugfix.

## [0.5.4] - 2020-12-08

### Fixed
- Fix telegram message layout.

## [0.5.3] - 2020-12-03

### Fixed
- Small fix.

## [0.5.2] - 2020-11-28

### Fixed
- Fix option variable.

## [0.5.0] - 2020-11-28

### Added
- Add supports for Tor hidden service.
- Add Debian package builder.
- Add logger.

### Changed
- Refactor code base.
- Refine packaging directory structure.

### Removed
- Remove debug mode of telegram-bot-api.

## [0.4.1] - 2020-11-12

### Added
- Handle request in parallel.
- Change default branch to main.
- Publish Docker images to GitHub Container Registry.

## [0.4.0] - 2020-10-16

### Changed
- Ending IPFS beta state.
- Refine Makefile.

## [0.3.2] - 2020-09-19

### Added
- Add dependabot config.

### Changed
- Upgrade to Go 1.15.

## [0.3.1] - 2020-08-31

### Fixed
- Fixed nil pointer.

## [0.3.0] - 2020-08-29

### Added
- Add build docker image workflows.
- Add cross compile target.
- Add linter workflows.

## [0.2.2] - 2020-08-23

### Added
- Add snapcraft badge.

### Fixed
- Fix release script.

## [0.2.0] - 2020-08-22

### Added
- Add snapcraft workflow.

## [0.1.0] - 2020-08-21

### Changed
- Refactor code base.

## [0.0.3] - 2020-07-25

### Changed
- Change IPFS default mode to pinner.

### Security
- Secure enhance for Tor.

## [0.0.2] - 2020-07-05

### Added
- Supports wayback to IPFS.

## [0.0.1] - 2020-07-05

### Added
- Initial release.
