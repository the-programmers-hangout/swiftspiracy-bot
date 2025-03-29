# swiftspiracy

A playful Discord bot that sends random praises and (occasionally) a conspiracy theory — and cleans up after itself.

## Setup

### Create a Discord Bot

> [!WARNING]
> Discord updates frequently -- if any steps seem outdated, refer to the official docs:
> https://discord.com/developers/docs/intro

1. Log into Discord.
2. Go to [Discord Developer Portal](https://discord.com/developers/applications)
3. Click "New Application", give it a name, and create it.
4. In the sidebar, go to Bot, click "Add Bot".
5. Under OAuth2 > URL Generator:
   1. Select "bot" scope.
   2. Under Bot Permissions, check at least:
      - Send Messages
6. Use the generated URL to invite the bot to your server.
7. In the Bot tab, click "Reset Token", copy the token and save it in your `.env` (view next step).

### Configuration

Create a `.env` file in the root of your project with the following keys:

- `DISCORD_BOT_TOKEN`: Your bot token from the Discord developer portal.
- `DISCORD_CHANNEL_ID`: The ID of the channel where the bot should send messages.

- `SEND_MESSAGE_INTERVAL_MIN`: Minimum interval between messages (as an integer).
- `SEND_MESSAGE_INTERVAL_MAX`: Maximum interval between messages (as an integer).
- `SEND_MESSAGE_UNIT`: Time unit for the interval. Acceptable values: "second", "seconds", "minute", "minutes",
  "millisecond", "milliseconds".

- `DELETE_CONSPIRACY_DELAY`: Time to wait before deleting a conspiracy message. Use Go duration format (eg "3s",
  "500ms", "1m").
- `CONSPIRACY_PROBABILITY`: Probability (from 0.0 to 1.0) that a conspiracy message is sent after a praise. Eg, 0.4 =
  40% chance.

You can view the [.env-sample](./.env-sample) as an example completed template.

### Install + Run

#### Vanilla (Local Dev)

1. Install Go via https://go.dev/doc/install
2. Install Make if you don't have it via https://www.gnu.org/software/make/
3. Run:
   ```bash
   make           # builds the bot
   ./bin/bot      # starts the bot
   ```

#### Docker

> We maintain a ghcr image at ghcr.io/the-programmers-hangout/swiftspiracy-bot:latest

1. Pull the image:
   ```bash
   docker pull ghcr.io/the-programmers-hangout/swiftspiracy-bot:latest
   ```
2. Run it with your `.env` file:
   ```bash
   docker run --env-file ./.env ghcr.io/the-programmers-hangout/swiftspiracy-bot:latest
   ```

## Notes

- The bot gracefully shuts down on CTRL+C.
- Conspiracy messages are deleted after a short delay to keep the mystery alive, you should configure it to be low.
- Messages are sourced from [praises.json](./cmd/bot/praises.json) and
  [conspiracies.json](./cmd/bot/conspiracies.json). If your timings for message are too short for the duration the bot
  runs, the messages will repeat.

### If using nix

*and if you are NOT using direnv*
At the project root directory, run:
`nix develop` and you'll have a development shell with the `go.mod` specified version of Golang (1.23.4 as of writing)
And if you're a chad, using direnv, you'll automatically load the environment when you `cd` into the project.

The flake also comes with **pre commit hooks** that will be run automatically when running `nix flake check`, it can also be extended to run arbitrary tests.
