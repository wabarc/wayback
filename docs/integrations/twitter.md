---
title: Interactive with Twitter
---

## How to build a Twitter Bot

Steps to create a new bot:

1. Create a Twitter account or use an existing one. Go to [signup](https://twitter.com/signup) if you need to create a new account.
2. Go to [Projects & Apps](https://developer.twitter.com/en/portal/projects-and-apps) and sign in with your Twitter account.
3. Click "Create an App" and fill in the required information such as the app name, description, and website.
4. On the "App details" page, click on the "Keys and tokens" tab.
5. Click the "Generate" button under the "Consumer Keys" section to generate the "Consumer Key" and "Consumer Secret" for your app.
6. Scroll down to the "Access token & access token secret" section and click the "Generate" button to generate the "Access Token" and "Access Token Secret" for your app.
7. Copy the "Consumer Key", "Consumer Secret", "Access Token", and "Access Token Secret" and store them in a secure location. You will need them later to authenticate your bot.
8. Place environment or configuration file for your bot, making sure to include the necessary authentication with the API using the keys and tokens generated in step 5 and 6.
9. Test your bot and make sure it is working as expected.
10. Once your bot is ready, deploy it to a hosting service or a server so it can run continuously.
11. Monitor your bot's activity and performance, and make any necessary adjustments.

Some things to keep in mind when creating a Twitter bot:

- Follow the Twitter rules and policies, as violating them can result in your bot being suspended or banned.
- Do not use your bot to spam or harass other Twitter users.
- Make sure your bot has a clear purpose and is not creating unnecessary noise on the platform.
- Monitor your bot's activity and make adjustments as needed to improve its performance and behavior.

## Configuration

After creating a new bot, you will have the `Consumer Keys`, `Consumer Secret`, `Access Token` and `Access Token Secret`.

Next, place these keys in the environment or configuration file:

- `WAYBACK_TWITTER_CONSUMER_KEY`: The customer key of your Twitter application.
- `WAYBACK_TWITTER_CONSUMER_SECRET`: The customer secret of your Twitter application.
- `WAYBACK_TWITTER_ACCESS_TOKEN`: The access token of your Twitter application.
- `WAYBACK_TWITTER_ACCESS_SECRET`: The access secret of your Twitter application.
