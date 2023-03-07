# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- markdownlint-disable -->
## [Unreleased]

### Added
- Add support for reads from stdin and files ([#301](https://github.com/wabarc/wayback/pull/301))
- Add support for publish to Nostr ([#311](https://github.com/wabarc/wayback/pull/311))
  - Message content styling
- Add documentation ([#330](https://github.com/wabarc/wayback/pull/330))

### Changed
- Sign images using cosign
- Replace set-output with recommended env output ([#234](https://github.com/wabarc/wayback/pull/234))
- Create deployment instructions for Render ([#236](https://github.com/wabarc/wayback/pull/236))
- Specify dependencies for the distribution package ([#243](https://github.com/wabarc/wayback/pull/243))
- Make media downloads are domain-specific ([#247](https://github.com/wabarc/wayback/pull/247))
- Always parse config file under daemon mode ([#271](https://github.com/wabarc/wayback/pull/271))
- Response uppercase letter for health check ([#292](https://github.com/wabarc/wayback/pull/292))
- Stores artifacts via screenshot ([#293](https://github.com/wabarc/wayback/pull/293))
- Improve signal handling ([#294](https://github.com/wabarc/wayback/pull/294))
- Improve httpd service ([#278](https://github.com/wabarc/wayback/pull/278))
  - Do not using pooling for http service
  - Only serve onion service with a valid torrc
  - Rename `HTTP_LISTEN_ADDR` to `WAYBACK_LISTEN_ADDR`
  - Support for `WAYBACK_LISTEN_ADDR` override `WAYBACK_TOR_LOCAL_PORT`
  - Defaults to listen `0.0.0.0` for httpd service
- Bump version for docker image ([#319](https://github.com/wabarc/wayback/pull/319))
  - Bump alpine to 3.17
  - Upgrade dependencies for docker workflow
  - No longer build image for `linux/s390x`

### Fixed
- Fix semgrep scan workflow ([#312](https://github.com/wabarc/wayback/pull/312))
- Fix terminal determination

## [0.18.1] - 2022-10-30

### Fixed
- Fix ipfs credential assignment ([#231](https://github.com/wabarc/wayback/pull/231),[#242](https://github.com/wabarc/wayback/pull/242))

### Changed
- Update repo url ([#241](https://github.com/wabarc/wayback/pull/241))
- Set the default path for the reduxer ([#235](https://github.com/wabarc/wayback/pull/235))
- Create pull_request_template.md ([#230](https://github.com/wabarc/wayback/pull/230))

## [0.18.0] - 2022-10-06

### Added
- Add support for placing a managed IPFS credential
- Add renovate.json ([#180](https://github.com/wabarc/wayback/pull/180))
- Add semgrep scan
- Add context cancellation for publish
- Add support for storing a page as a single file ([#184](https://github.com/wabarc/wayback/pull/181))
- Add support for push documents to Meilisearch ([#174](https://github.com/wabarc/wayback/pull/174))
- Add support for publishing to notion
- Add support for retrying wayback requests
- Add retry strategy for publish
- Add support for installing from Bina

### Changed
- Improve reduxer calls
- Enable all wayback slot
- Upload packages to Gemfury ([#223](https://github.com/wabarc/wayback/pull/223))
- Add testing for config
- Set go version to 1.19 for build binary
- Upgrade dependencies
- Run golangci-lint on multiple os
- Turns the pooling bucket into a non-pointer
- Remove unused code from pooling
- Handle startHTTPServer with goroutine
- Set up the Meilisearch server for testing workflow
- Minor improvements to the service goroutine
- Context leak detection
- Meilisearch endpoint version compatible ([#185](https://github.com/wabarc/wayback/pull/185))
- skywalking-eyes now has a dedicated header checker path ([#181](https://github.com/wabarc/wayback/pull/181))
- Minor enhancements to the worker pool
- Pin non-official workflow dependencies
- Upload coverage to Codecov
- Cache go module for workflow
- Upgrade the go version for the Docker workflow
- Removing the retry strategy for publishing
- Minor improvements for processing notion block
- Several improvements from `telegra.ph`
- Minor changes for render testing
- Convert the publish flag to a name

### Fixed
- Fix install command to use `go install`
- Fix license checker
- Fix markdown link
- Fix golang linter
- Fix testing workflow
- Fix unspecified failure response message

## [0.17.0] - 2022-03-14

### Added
- Add exempt rules for stale workflow
- Add FOSSA Action
- Add license checker workflow
- Add lock for pooling
- Add install script
- Supports profiling in debug mode
- Transform telegram message entities

### Changed
- Upgrade base image to Alpine 3.15
- Wayback to IPFS with bundled HTML
- Converting byte slice and string without memory allocation
- Rename package iawia002/annie to iawia002/lux
- Upgrade go directive in go.mod to 1.17
- Store resources to IPFS from a directory
- Upgrade tucnak/telebot to v3
- Remove duplicates url
- Backward compatibility systemd with windows
- Handle download media outputs
- Refine permissions for codeql actions
- Refine reduxer bundle
- Change the pooling to a pointer
- Minor improvement for reduxer
- Move upload funcs to service utils
- Request final URI before wayback
- Refactoring of reduxer
- Refine metrics constant
- Build snapcraft using snapcore/action-build
- Bump actions/checkout from 2 to 3
- Bump actions/\* from v2 to v3
- Upgrade dependencies
- Use go 1.18 for testing

### Fixed
- Fix testing
- Fixed cannot publish to telegram channel from other services
- Closes response body to fix go lint
- Fix data race in reduxer
- Unset specified env to make actions green

## [0.16.2] - 2021-12-04

### Added
- Add wayback user agent
- Add header parameters for warcraft
- Add an option to enable URL fallback

### Changed
- Make wayback to IPFS as default
- Build docker image for develop branch
- Enhancements for youtube-dl media downloads
- Dispatch repository in wabarc/homebrew-wayback
- Change the URL fallback defaults to disabled and enable it with the `WAYBACK_FALLBACK` environment variable.
- Increase the worker pool timeout to more than 3 seconds
- Set the user agent for the download of the warc file
- Download media with specific format
- Minor improvement for render assets url
- Minor improvements in testing
- Upgrade dependencies

### Fixed
- Improvement for create warc file
- Fix wget warc parse error

## [0.16.1] - 2021-10-24

### Fixed
- Fix releasing binaries for windows are missing

## [0.16.0] - 2021-10-24

### Added
- Add support for export HAR file
- Add specific permissions to workflows under .github/workflows
- Supports to close worker pool
- Starts http service as clear web if missing tor
- Gracefully shuts down services
- Add support for systemd (#110)

### Changed
- Releasing defaults to pre-release
- Refine testing workflow
- Improvements for golint
- Upgrade Go version to 1.17
- Minor improvement for worker pool
- Makes silent for downloading media via Annie
- Makes wayback timeout configurable
- Update Tor socks port default to 9050
- Refine makefile (#111)

### Fixed
- Fix nil pointer dereference if `WAYBACK_STORAGE_DIR` not set
- Check received content for testing
- Fix httpd service's playback gauge record to wayback
- Fix worker pool
- Fix data race for discord testing

## [0.15.1] - 2021-08-13

### Changed
- Handle debugging message from tucnak/telebot
- Upgrade dependencies

### Fixed
- Fix docker tag
- Fix pooling scalable

## [0.15.0] - 2021-08-05

### Added
- Add support for Slack
- Add support for Discord
- Add support for download stream media
- Bundle all requirements in one image
- Upload files remotely for sharing
- Supports to serve text content

### Changed
- Minor improvements
- Download media via you-get
- Download media via youtube-dl
- Support for replying to message from group/channel and mention bot to wayback
- Apply logger color
- Use Fedora 34 to build RPM package
- Use parallel flag for testing
- Minor improvements for readability
- Minor improvements
- Refine logger message
- Format output for print configurations
- Add timeout for wayback context
- Bump actions/stale from 3 to 4

## [0.14.1] - 2021-07-12

### Changed
- Styling outputs and message
- Print stored files for cmd
- Refine returns value for archive.org
- Strip blank node for telegra.ph

### Fixed
- Fix tests

## [0.14.0] - 2021-07-07

### Added
- Summarize for publish and readability content for Telegra.ph
- Add support to serve WARC file
- Add Sonatype Nancy to check for vulnerabilities
- Attaching hashtag to the Mastodon toot

### Changed
- Minor improvements: waitgroup => errgroup
- Standardize the description of Docker images
- Disable to releasing snap if pull requests
- Refactor: publish multiple message
- Improvement for web layout
- Misc updates

## [0.13.1] - 2021-06-27

### Added
- Add publish to telegram private channel support

### Changed
- Improvements for playback ([wabarc/playback](https://github.com/wabarc/playback/commit/d3f173eb76b2eca0ed2fcbbaae24778bce0064ef))
- Extract title from reduxer bundles
- Set environment from wayback.conf automatically
- Set env for testing and refine workflows
- Improve some code

### Fixed
- Do not publish playback results from web request

## [0.13.0] - 2021-06-19

### Added
- Add support store archived files to disk
- Supports playback for web, mastodon and matrix
- Supports playback from google cache
- Supports mention from Mastodon
- Packaging Flatpak and Snapcraft
- Add heroku one click deploy button

### Changed
- Replace service/anonymity to service/httpd
- Change onion service address
- Refine some code & improve post tweet
- Extract title for github issue
- Chore changes

### Fixed
- Fix linter

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

<!-- markdownlint-restore -->
