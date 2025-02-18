# Your personal endurance coach
Roachy is your personal endurance coach, you can connect him to Strava and Discord and it will guide you through your training with actionable insights, solid roasting and some deep science-based tips.

This is using an experimental feature of [Dagger](https://dagger.io) that provides native LLM support, integrating the APIs of Dagger Modules in an LLM conversation via tool calling. For this application I developed a module [athlete-workspace](/athlete-workspace) that wraps two separate modules into one:
- [Strava](/strava): wrapper to calling Strava's API and fetch activities, group activities and other athlete information.
- [Notify](https://github.com/gerhard/daggerverse/tree/main/notify): wrapper for calling Discord's API and send messages to a channel.

Via these two modules I can create a conversation with Roachy, where I can provide an activity ID and it will take care of analysing the activity in the context of the athlete and team, providing insights and good roasting. Demo:

![demo](https://bunny.matiaspan.dev/roachy-demo.gif)

## Features

### Roasting you and your friends

Connect your Strava webhooks with Roachy and every time it receives an activity it will roast you and your friends on Discord, making sure that it keeps track of who is leading in the team and giving a hard time for everybody falling behind.

