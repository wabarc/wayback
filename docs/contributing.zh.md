# 贡献

你好！非常高兴看到你想为这个项目做出贡献。你的帮助对于保持项目的优秀非常重要。

请注意，这个项目发布了[贡献者行为准则][code-of-conduct]。参与这个项目的过程中，你同意遵守这个准则。

## 发布新版本

我们使用[semantic-release](https://github.com/semantic-release/semantic-release)来自动发布新版本。

一次只会提升一个版本号，最高版本的更改会覆盖其他版本的更改。除了将 Docker 镜像发布到 Docker Hub 和 GitHub Packages，semantic-release 还会在 GitHub 上创建 Git 标签和发布，生成版本二进制文件摘要并将其放入发布说明中。

## 提交 Pull Requests

1. [Fork][fork] 并克隆仓库。
2. 确保你的机器上测试通过：`make test` 或 `go test -v ./...`
3. 创建新分支：`git checkout -b my-branch-name`
4. 进行更改、添加测试，并确保测试仍然通过。
5. 推送到你的 Fork 上并[提交 Pull Requests][pr]。
6. 为自己鼓掌，等待你的 Pull Request 被审查并合并。

我们也欢迎 Work in Progress 的 Pull Requests，以便你可以尽早得到反馈，或者如果你有什么困难。

## 报告 Bug

### 先决条件

创建 Bug 报告时，最重要的细节是确定是否真的需要创建一个 Bug 报告。

#### 先做研究

你是否研究过现有的问题、已关闭和开放的问题，看看其他用户是否遇到并可能已经解决了你遇到的相同问题？

#### 描述问题，不要轻易得出结论

最近，我和另一位维护者讨论了这个话题。我们想知道我们处理了多少个标题中包含 "Bug" 这个词或类似的内容的问题，最终发现是用户错误或绝对不是 Bug。这只是一个猜测，但我认为说只有 10 个标题中带有 "Bug" 的报告中，只有 1 个最终是 Bug。

### 重要细节

当需要 Bug 报告时，绝大多数的 Bug 报告应该包括以下四个信息：

1. `版本`
2. `描述`
3. `错误信息`
4. `代码`

## 功能请求

在提交功能请求之前，请尝试熟悉项目。了解项目是否有特定的目标或指导方针，描述功能请求的方式应该是怎样的。

## 刚开始？寻找如何帮助？

使用[这个搜索工具][good-first-issue-search]查找已标记为 `good-first-issue` 的 Wayback Archiver 问题。

## 资源

- [如何贡献开源项目](https://opensource.guide/how-to-contribute/)
- [使用 Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub 帮助](https://help.github.com)

[fork]: https://github.com/wabarc/wayback/fork
[pr]: https://github.com/wabarc/wayback/compare
[code-of-conduct]: ./CODE_OF_CONDUCT.md
[good-first-issue-search]: https://github.com/search?q=org%3Awabarc+good-first-issues%3A%3E0
