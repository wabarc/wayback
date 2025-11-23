---
title: Matrix
---

![Matrix Room](../assets/matrix-room.png)

## 如何构建Matrix机器人

您可以选择任何Matrix服务器。这里，我们将使用**matrix.org**和**Element**作为示例。

要注册Matrix帐户，请按照以下步骤操作：

1. 打开[Element](https://app.element.io/)并单击“创建帐户”。
2. 填写所需的信息。
3. 登录并创建一个**公共房间**以发布（可选）。
4. 转到**房间设置**>**高级**，您可以找到**内部房间ID**（可选）。

## 配置

创建Matrix帐户后，您将拥有`Homeserver`、`User ID`、`Password`和`Internal room ID`。

接下来，将这些密钥放置在环境或配置文件中：

- `WAYBACK_MATRIX_HOMESERVER`：您选择的Homeserver，默认为`matrix.org`
- `WAYBACK_MATRIX_USERID`：用户ID，例如`@alice:matrix.org`
- `WAYBACK_MATRIX_ROOMID`：内部房间ID
- `WAYBACK_MATRIX_PASSWORD`：从您的注册步骤中获取的密码。

## 相关资料

- [开发人员指南](https://matrix.org/docs/develop/)
