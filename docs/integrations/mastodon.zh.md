---
title: Mastodon
---

![Mastodon](../assets/mastodon.png)

## 如何构建Mastodon机器人

您可以选择任何Mastodon实例。这里，我们将使用Mastodon.social作为示例。

要创建Mastodon应用程序，您可以按照以下步骤操作：

1. 登录到您的Mastodon帐户。
2. 转到“设置”>“[开发](https://mastodon.social/settings/applications)”>“[新应用程序](https://mastodon.social/settings/applications/new)”。
3. 输入以下信息：
   - **应用程序名称**：您的应用程序名称。
   - **应用程序网站**：与您的应用程序相关联的网站。
   - **重定向URI**：用户在身份验证后将被重定向的URI。这可以是任何有效的URI，但它必须与您的代码中指定的重定向URI匹配。
   - 范围：您的应用程序需要的范围。这些范围确定您的应用程序可以代表用户执行的操作。所需的最小范围为：`read:statuses`、`read:notifications`、`write:statuses`、`write:notifications`和`write:conversations`。
4. 单击“提交”。
5. 在下一页上，您将看到您的应用程序的客户端ID和客户端密钥。这些将用于验证您的应用程序。

您可以在Mastodon文档中找到有关创建Mastodon应用程序的更多信息：https://docs.joinmastodon.org/client/token/

## 配置

创建Mastodon应用程序后，在应用程序详细信息页面上可以找到`Client key`、`Client secret`和`Your access token`。

接下来，将这些密钥放置在环境或配置文件中：

- `WAYBACK_MASTODON_KEY`：客户端密钥
- `WAYBACK_MASTODON_SECRET`：客户端密钥
- `WAYBACK_MASTODON_TOKEN`：您的访问令牌

另外，您必须通过设置`WAYBACK_MASTODON_SERVER`变量来指定Mastodon服务器。

## 相关资料

- [Fediverse Observer](https://mastodon.fediverse.observer/list)
