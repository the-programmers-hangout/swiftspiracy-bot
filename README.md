# swiftspiracy

A playful Discord bot that sends random praises and (occasionally) a conspiracy theory — and cleans up after itself.

## Setup

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

You can view the [.env-sample](./.env-sample) as an example.

## Notes

- The bot gracefully shuts down on CTRL+C.

### If using nix

*and if you are NOT using direnv*
At the project root directory, run:
`nix develop` and you'll have a development shell with the `go.mod` specified version of Golang (1.23.4 as of writing)
And if you're a chad, using direnv, you'll automatically load the environment when you `cd` into the project.

The flake also comes with **pre commit hooks** that will be run automatically when running `nix flake check`, it can also be extended to run arbitrary tests.
