---
title: Interactive with Web Service
---

## How to build a Web Service

Wayback supports serving both **Clear Web** and **Onion Service**. If the Tor binary is missing, the Onion Service feature will be ignored.

## Configuration

After installation, you need to provide the required keys by placing them in the environment or configuration file. This allows you to customize the configuration based on your needs.

- `WAYBACK_LISTEN_ADDR`: The listen address for the HTTP server, defaults to `0.0.0.0:8964`.
- `WAYBACK_ONION_PRIVKEY`: The private key for Onion Service.
- `WAYBACK_ONION_LOCAL_PORT`: Local port for Onion Service, also support for a reverse proxy. This is ignored if `WAYBACK_LISTEN_ADDR` is set.
- `WAYBACK_ONION_REMOTE_PORTS`: Remote ports for Onion Service, e.g. `WAYBACK_ONION_REMOTE_PORTS=80,81`.

Note: To run a Onion Service for the first time, you need to keep the `private key`, which can be seen from the log output.
