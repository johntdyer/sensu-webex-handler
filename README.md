# Sensu Go webex Handler

The Sensu webex handler is a [Sensu Event Handler][1] that sends event data to
a configured webex channel.   This plugin was mostly copied from the [Sensu slack handler][2] and repurposed for [Webex Teams][4].

![screenshot](images/example.png?raw=true "Example")

## Installation

Download the latest version of the sensu-webex-handler from [releases][2],
or create an executable script from this source.

From the local path of the webex-handler repository:
```
go build -o /usr/local/bin/sensu-webex-handler main.go
```

## Configuration

Example Sensu Go handler definition:

webex-handler.json

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "webex"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-webex-handler --token abc123 --room-id 'ABCDEFGHIJKLMNOP123' --timeout 20 \\",
        "timeout": 30,
        "filters": [
            "is_incident"
        ]
    }
}
```

`sensuctl create -f webex-handler.json`

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
            "webex"
        ]
    }
}
```

## Usage examples

Help:

```
The Sensu Go webex handler for notifying a channel

Usage:
  sensu-webex-handler [flags]

Flags:
  -c, --room-id string       The space to post messages to, can also be a users email if you want to send directly to a person vs a space
  -t, --token string         Api token to use
  -h, --help                 help for handler-webex
  -t, --timeout int          The amount of seconds to wait before terminating the handler (default 10)
  
```

[1]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[2]: https://github.com/johntdyer/sensu-webex-handler/releases
[3]: https://github.com/sensu/sensu-slack-handler
[4]: https://developer.webex.com