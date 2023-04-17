## archive.today不可用？

有时，archive.today会实施严格的验证码策略，这可能会导致请求异常。为了解决这个问题，您可以在浏览器中手动绕过验证码并检索`cf_clearance` cookie项。然后，将此项设置为名为`ARCHIVE_COOKIE`的系统环境变量。例如，您可以将`ARCHIVE_COOKIE`的值设置为`cf_clearance=ab170e4acc49bbnsaff8687212d2cdb987e5b798-1234542375-KDUKCHU`。

## 在使用IPFS存档时如何为特定的URI禁用JavaScript？

要在保存网页时禁用JavaScript，您可以设置环境变量`DISABLEJS_URIS`。值应按照以下格式：

```sh
export DISABLEJS_URIS=wikipedia.org|eff.org/tags
```

这将为整个网站`wikipedia.org`禁用JavaScript，或者仅在匹配时为路径`eff.org/tags`禁用。

## 如何保留Tor隐藏服务主机名？

第一次运行`wayback`服务时，保留来自输出消息的私钥非常重要（私钥是`private key:`之后的部分）。下次运行wayback服务时，您可以通过将其提供给`--tor-key`选项或将其设置为`WAYBACK_ONION_PRIVKEY`环境变量来使用密钥。

```text
[INFO] Web: Important: remember to keep the private key: d005473a611d2b23e54d6446dfe209cb6c52ddd698818d1233b1d750f790445fcfb5ece556fe5ee3b4724ac6bea7431898ee788c6011febba7f779c85845ae87
```

