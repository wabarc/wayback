---
title: Interactive with Matrix
---

## How to build a Matrix Bot

You can choose any Matrix server. Here, we will be using **matrix.org** and **Element** as an example.

To register a Matrix account, follow these steps:

1. Open [Element](https://app.element.io/) and click "Create Account".
2. Fill in the required information.
3. Log in and create a **public room** for publishing (optional).
4. Go to **Room Settings** > **Advanced**, you can find **Internal room ID** (optional).

## Configuration

After creating a Matrix account, you will have the `Homeserver`, `User ID`, `Password`, and `Internal room ID`.

Next, place these keys in the environment or configuration file:

- `WAYBACK_MATRIX_HOMESERVER`: Homeserver of your choice, defaults to `matrix.org`
- `WAYBACK_MATRIX_USERID`: User ID, e.g. `@alice:matrix.org`
- `WAYBACK_MATRIX_ROOMID`: Internal room ID
- `WAYBACK_MATRIX_PASSWORD`: Password from your registration step.

## Further reading

- [Guides for Developers](https://matrix.org/docs/develop/)
