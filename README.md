# Simple Telegram forwarder

A small tool for automatically forwarding telegram messages from source chats to destination chats in real time as messages appear.

- Monitoring multiple sources and forwarding/copying messages to all the defined destinations
- Forwarding (with 'Forwarded to' label) or copying (posting as new) messages
- Filtering messages with regular expressions, so only messages that match the filter get copied/forwarded
- `--auth-only` flag to only interactively login to Telegram and then exit
- The app uses Telegram Client API, so there is no need to create any bots and be an admin of the group/channel from where you
  want to forward messages
- You don't need to trust any third party software or service with your Telegram account, this app
  does not force you to forward your Telegram OTP code anywhere.

### Usage

You'll need a Telegram account for the app to work. The account must be a member and
have necessary permissions in the chats/groups/channels between which you want to forward messages.

- [Create an application](https://core.telegram.org/api/obtaining_api_id)
- Obtain your `api_id` and `api_hash`
- [Create a configuration](#configuration) JSON file
- Run as a [console application](#building-and-running-an-executable) or with [Docker](#running-with-docker)
    - When run for the first time, follow console messages to log in to your account
    - Optionally, run with `--auth-only` flag to only authorize to Telegram and then exit

### Configuration

Configuration is done via a json file. File name of the json file can be specified with `CONFIG_FILE` environment variable, by default it
is `simple-telegram-forwarder.config.json`.

| Field JSONPath                                 | Example value                    | Description                                                                                                                     |
|------------------------------------------------|----------------------------------|---------------------------------------------------------------------------------------------------------------------------------|
| `$.api_hash`                                   | `cb01sdfe7922afd2970359f6aa76d0` | `api_hash` obtained from creating a Telegram application                                                                        |
| `$.api_id`                                     | `98011131`                       | `api_id` obtained from creating a Telegram application                                                                          |
| `$.forwarding_config.sources[*].username`      | `@telegram`                      | Username or channel name of the source. Use this field or `$.forwarding_config.sources[*].chat_id`                                  |
| `$.forwarding_config.sources[*].chat_id`       | `-1001005640892`                 | Chat ID of the source. Use i.e. `@userinfobot` to get it. Use this field or `$.forwarding_config.sources[*].username`               |
| `$.forwarding_config.destinations`             |                                  | Array of destinations. Must contain at least one                                                                                |
| `$.forwarding_config.destinations[*].username` | `@telegram`                      | Username or channel name of the source. Use this field or `$.forwarding_config.destinations[*].chat_id`                         |
| `$.forwarding_config.destinations[*].chat_id`  | `-1001005640892`                 | Chat ID of the destination. Use i.e. `@userinfobot` to get it. Use this field or `$.forwarding_config.destinations[*].username` |
| `$.forwarding_config.forward`                  | `bool`                           | Forward messages instead of sending a copy. Default: false                                                                      |
| `$.forwarding_config.filter`                   | `(?i)(any\|regex?\|(you)*want)`  | Optional regular expression for message filtering: only matched messages are forwarded.                                         |

Example configuration file:

```json
{
  "api_hash": "cb01sdfe7922afd2970359f6aa76d0",
  "api_id": 98011131,
  "forwarding_config": {
    "sources": [{
      "username": "@telegram"
    },{
      "chat_id": "-1001005640893"
    }],
    "destinations": [
      {
        "chat_id": -1001005640892
      }
    ],
    "forward": true,
    "filter": {
      "regex": "(?i)(any|regex?|(you)*want)"
    }
  }
}
```

### Building and running an executable

In order to compile and run this application you'll need a TDLib library installed on your system. Please refer
to TDLib [building instructions](https://github.com/tdlib/td?tab=readme-ov-file#building).
Then run:

```shell
go build .
./simple-telegram-forwarder
```

If it is the first run, then you will be prompted to enter your phone and OTP code.
After that TDLib will create `.tdlib` folder in the working directory which stores session data.

If you wish to decouple logging in to Telegram from normal app usage, then run the executable with `--auth-only` flag. 

```shell
./simple-telegram-forwarder --auth-only
```

### Running with Docker

Build docker image (may take a while):
```shell
docker build -t simple-telegram-forwarder .
```

Or pull an existing one from DockerHub (for linux-amd64 only):
```shell
docker pull marchukd/simple-telegram-forwarder
```

When running for the first time, pass`--auth-only` flag to make application interactively authorize to Telegram and then exit:
```shell
docker container run -it -v ./.tdlib:/app/.tdlib -v ./simple-telegram-forwarder.config.json:/app/simple-telegram-forwarder.config.json:ro simple-telegram-forwarder ./simple-telegram-forwarder --auth-only
```

After that execute to run the app normally:
```shell
docker run -v ./.tdlib:/app/.tdlib -v ./simple-telegram-forwarder.config.json:/app/simple-telegram-forwarder.config.json:ro simple-telegram-forwarder
```

