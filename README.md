# Sensu Go Slack Handler

The Sensu Slack handler is a [Sensu Event Handler][1] that sends event data to
a configured Slack channel.

## Installation

Download the latest version of the sensu-slack-handler from [releases][2],
or create an executable script from this source.

From the local path of the slack-handler repository:
```
go build -o /usr/local/bin/sensu-slack-handler main.go
```

## Configuration

Example Sensu Go handler definition:

slack-handler.json

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "slack"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-slack-handler --channel '#general' --timeout 20 --username 'sensu' --webhook-url 'https://www.webhook-url-for-slack.com'",
        "timeout": 30,
        "filters": [
            "is_incident"
        ]
    }
}
```

`sensuctl create -f slack-handler.json`

Example Sensu Go check definition:

```json
{
    "api_version": "core/v2",
    "type": "CheckConfig",
    "metadata": {
        "namespace": "default",
        "name": "dummy-app-healthz"
    },
    "spec": {
        "command": "check-http -u http://localhost:8080/healthz",
        "subscriptions":[
            "dummy"
        ],
        "publish": true,
        "interval": 10,
        "handlers": [
            "slack"
        ]
    }
}
```

## Usage examples

Help:

```
The Sensu Go Slack handler for notifying a channel

Usage:
  sensu-slack-handler [flags]

Flags:
  -c, --channel string       The channel to post messages to (default "#general")
  -h, --help                 help for handler-slack
  -i, --icon-url string      A URL to an image to use as the user avatar (default "http://s3-us-west-2.amazonaws.com/sensuapp.org/sensu.png")
  -t, --timeout int          The amount of seconds to wait before terminating the handler (default 10)
  -u, --username string      The username that messages will be sent as (default "sensu")
  -w, --webhook-url string   The webhook url to send messages to
```

[1]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[2]: https://github.com/sensu/sensu-slack-handler/releases
