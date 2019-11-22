# Sensu Go webex Handler [![Build Status](https://travis-ci.com/johntdyer/sensu-webex-handler.svg?branch=master)](https://travis-ci.com/johntdyer/sensu-webex-handler)

[[!GitHub issues](https://img.shields.io/github/issues/johntdyer/sensu-webex-handler)]
[[!GitHub forks](https://img.shields.io/github/forks/johntdyer/sensu-webex-handler)]
[[!GitHub stars](https://img.shields.io/github/stars/johntdyer/sensu-webex-handler)]
[[!GitHub license](https://img.shields.io/github/license/johntdyer/sensu-webex-handler)]

The Sensu webex handler is a [Sensu Event Handler][1] that sends event data to
a configured webex channel.   This plugin was mostly copied from the [Sensu slack handler][2] and repurposed for [Webex Teams][4].

![screenshot](images/cards-example.png?raw=true "Example")

## Installation

Download the latest version of the sensu-webex-handler from [releases][2],
or create an executable script from this source.

From the local path of the webex-handler repository:

```go
go build -o /usr/local/bin/sensu-webex-handler main.go template.go
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
        "command": "sensu-webex-handler --api-key abc123 --room-id 'ABCDEFGHIJKLMNOP123' \\",
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

```shell
The Sensu Go Webex Teams handler for event forwarding

Usage:
  sensu-webex-teams-handler [flags]

Flags:
  -k, --api-key string   API key for authenticated access
  -a, --api-url string   Webex Teams API (default "api.ciscospark.com")
  -h, --help             help for sensu-webex-teams-handler
  -r, --room-id string   Room or email to send alert to

```

[1]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[2]: https://github.com/johntdyer/sensu-webex-handler/releases
[3]: https://github.com/sensu/sensu-slack-handler
[4]: https://developer.webex.com
