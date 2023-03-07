## archive.today is unavailable?

Sometimes archive.today enforces a strict CAPTCHA policy, which may cause a request exception.
To solve this, you can manually bypass the CAPTCHA in your browser and retrieve the `cf_clearance` cookie item.
Then, set this item as a system environment variable named `ARCHIVE_COOKIE`. For example, you can set the value
of `ARCHIVE_COOKIE` as `cf_clearance=ab170e4acc49bbnsaff8687212d2cdb987e5b798-1234542375-KDUKCHU`.

## Disable JavaScript for specific URIs when archiving with IPFS?

To disable JavaScript when saving a webpage, you can set environment variables `DISABLEJS_URIS`. The values should be in the following format:

```sh
export DISABLEJS_URIS=wikipedia.org|eff.org/tags
```

This will disable JavaScript for the entire site `wikipedia.org` or only for the path `eff.org/tags` if it matches.

## How can I keep the Tor Hidden Service hostname?

When running the `wayback` service for the first time, it is important to keep the private key from the output message
(the key is the part after `private key:`). The next time you run the wayback service, you can use the key by providing
it to the `--tor-key` option or setting it as the `WAYBACK_TOR_PRIVKEY` environment variable.

```text
[INFO] Web: Important: remember to keep the private key: d005473a611d2b23e54d6446dfe209cb6c52ddd698818d1233b1d750f790445fcfb5ece556fe5ee3b4724ac6bea7431898ee788c6011febba7f779c85845ae87
```

