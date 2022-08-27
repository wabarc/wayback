# Contributing

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Please note that this project is released with a [Contributor Code of Conduct][code-of-conduct]. By participating in this project you agree to abide by its terms.

## Releasing new version

Releases are automated using [semantic-release](https://github.com/semantic-release/semantic-release).

Only one version number is bumped at a time, the highest version change trumps the others.
Besides publishing a Docker image to Docker Hub and GitHub Packages, semantic-release also
creates a git tag and release on GitHub, generates digests of the release binaries and puts
them into the release notes.

## Submitting Pull Requests

1. [Fork][fork] and clone the repository
1. Make sure the tests pass on your machine: `make test` or `go test -v ./...`
1. Create a new branch: `git checkout -b my-branch-name`
1. Make your change, add tests, and make sure the tests still pass
1. Push to your fork and [submit a pull request][pr]
1. Pat your self on the back and wait for your pull request to be reviewed and merged.

Work in Progress pull requests are also welcome to get feedback early on, or if there is something that blocked you.

## Reporting Bugs

### Prerequisites

The most important detail to consider when creating a bug report is whether or not you should create one at all.

#### Do research first

Did you research existing issues, closed and open, to see if other users have experienced
(and potentially already solved) the same issue you're having?

#### Describe the problem, don't jump to conclusions

Another maintainer and I were discussing this topic recently. We wondered how many issues
we've handled that were created with the word "bug" in the title, or something along those
lines that ended up being user error or were definitely not a bug. This is a guesstimate,
but I think it's conservative to say that only 1 out of 10 reports with "bug" in the title
has actually ended up being a bug.

### Important Details

When a bug report is warranted, the vast majority of bug reports should include the following
four bits of information:

1. `version`
1. `description`
1. `error messages`
1. `code`

## Feature Requests

Before submitting a feature request, try to get familiarized with the project. Find out if the
project has certain goals, or guidelines that describe how feature requests should be made.

## Just starting out? Looking for how to help?

Use [this search][good-first-issue-search] to find Wayback Archiver that have issues marked with the `good-first-issue` label.

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)

[fork]: https://github.com/wabarc/wayback/fork
[pr]: https://github.com/wabarc/wayback/compare
[code-of-conduct]: ./CODE_OF_CONDUCT.md
[good-first-issue-search]: https://github.com/search?q=org%3Awabarc+good-first-issues%3A%3E0
