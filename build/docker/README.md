## How do I verify images?

All images are signed by [cosign](https://github.com/sigstore/cosign). We recommend verifying any wayback image you use.

Once you've installed cosign, you can use the [wayback public key](https://github.com/wabarc/wayback/blob/main/cosign.pub) to verify any wayback image with:

```bash
$ cat cosign.pub
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAET+oZJBKR2xJF6Jj6yH7rEL+5/LP4
jtCZHNwHP1b1CP1mWRIFjzSZbEo0/4ZopFHs2d5qNDbphvCXI6gjEZHmnw==
-----END PUBLIC KEY-----

$ cosign verify --key cosign.pub $IMAGE_NAME
```

