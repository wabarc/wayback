# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add Mastodon support.
- Supports publish toot even if the entry from Telegram Bot and Tor Hidden Service.
- Add Twitter support.
- Supports publish tweet even if the entry from Mastodon Bot, Telegram Bot and Tor Hidden Service.
- Add stale workflow.

### Changed
- Make logs more reaadable.
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
