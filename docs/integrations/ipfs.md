---
title: Wayback to IPFS
---

Wayback relies on the InterPlanetary File System ([IPFS](https://ipfs.tech/)) as an upstream service to store complete web pages,
including all related assets like JavaScript, CSS, and fonts. This allows for seamless playback of archived web pages,
ensuring that the user experience is identical to the original site.

You can enable or disable this feature using the `--ip` flag or the `WAYBACK_ENABLE_IP` environment variable, which is enabled by default.

The code for wayback's implementation of the Internet Archive integration can be found in the [wabarc/rivet](https://github.com/wabarc/rivet) repository.
