---
title: Publish to Notion
---

## How to build a Notion Bot

1. Sign up for a Notion account, if you don't have one already.
2. Create a new Integration by going to the [Notion API page](https://www.notion.so/my-integrations) and clicking on "**My integrations**" in the top-right corner of the page.
3. Give your integration a name and click on "**Create Integration**".
4. On the integration page, click on "**Add a new integration**".
5. Select "**Internal Integration**" and click "**Submit**".
6. On the next page, you will see your "**Integration Token**". Copy this token as you will need it later.
7. Grant your integration access to a database by sharing the database with your integration. To do this, go to the database you want to use with your bot, click on the three-dot menu, and select "**Share**".
8. In the "**Connections**" field, navigate to "**Add connections**", find your integration and select it from the list.
9. Select the appropriate permissions for your integration and click on "**Confirm**".

Note: A Notion database must be created using the "**New page**" option, not the "**Add a page**" option.

## Configuration

After creating a new account, you will have the `Integration Token` and `Notion database ID`.

Next, place these keys in the environment or configuration file:

- `WAYBACK_NOTION_TOKEN`: Notion integration token.
- `WAYBACK_NOTION_DATABASE_ID`: Notion database ID for archiving results.

