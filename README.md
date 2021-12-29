# Telegram Bot for Grafana's Annotations [![Build Status](https://cloud.drone.io/api/badges/13rentgen/grafana-annotations-bot/status.svg)](https://cloud.drone.io/13rentgen/grafana-annotations-bot)

This is the [Grafana annotations](http://docs.grafana.org/http_api/annotations/) Telegram bot that notifies you when new annotations will be added to Grafana.  

## Commands

###### /start tagName,anotherOneTag

> You're subscribe for tags:
> tagName
> anotherOneTag  

###### /stop

> You're successfully unsubscribe for tags:
> tagName
> anotherOneTag

###### /status

> **Grafana**
> Version: 5.4.2
> Database: ok
> 
> **Telegram Bot**
> Version: go1.11.4
> Uptime: Thu, 21 Mar 2019 11:45:24 MSK

## Installation

### Build from source

`go get github.com/13rentgen/grafana-annotations-bot`

### Configuration
| Flag                             | ENV                              | Required | Default                | Description                                                                                             |
|----------------------------------|----------------------------------|----------|------------------------|---------------------------------------------------------------------------------------------------------|
| --grafana.url                    | GRAFANA_URL                      | True     |                        | The URL that's used to connect to the Grafana, example: `http://localhost:3000`                         |
| --grafana.token                  | GRAFANA_TOKEN                    | True     |                        | The Bearer token used to connect with Grafana API                                                       |
| --grafana.scrapeInterval         | GRAFANA_SCRAPE_INTERVAL          | False    | `10s`                  | Scrape annotations interval                                                                             |
| --grafana.tls.insecure           | GRAFANA_TLS_INSECURE             | False    | `false`                | Insecure connection to Grafana API                                                                      |
| --grafana.tls.insecureSkipVerify | GRAFANA_TLS_INSECURE_SKIP_VERIFY | False    | `false`                | Grafana TLS config - insecure skip verify                                                               |
| --grafana.tls.cert               | GRAFANA_TLS_CERT                 | False    |                        | Grafana TLS config - client cert file path                                                              |
| --grafana.tls.key                | GRAFANA_TLS_KEY                  | False    |                        | Grafana TLS config - client key file path                                                               |
| --store.type                     | STORE_TYPE                       | False    | `bolt`                 | The store to use. Possible values: `bolt`, `etcd`                                                       |
| --store.keyPrefix                | STORE_KEY_PREFIX                 | False    | `annotationsbot/chats` | Prefix for store keys                                                                                   |
| --bolt.path                      | BOLT_PATH                        | False    | `/tmp/bot.db`          | Bolt database file path                                                                                 |
| --etcd.url                       | ETCD_URL                         | False    | `localhost:2379`       |                                                                                                         |
| --etcd.tls.insecure              | ETCD_TLS_INSECURE                | False    | `false`                | Insecure connection to ETCD                                                                             |
| --etcd.tls.insecureSkipVerify    | ETCD_TLS_INSECURE_SKIP_VERIFY    | False    | `false`                | ETCD TLS config - insecure skip verify                                                                  |
| -etcd.tls.cert                   | ETCD_TLS_CERT                    | False    |                        | ETCD TLS config - client cert file path                                                                 |
| --etcd.tls.key                   | ETCD_TLS_KEY                     | False    |                        | ETCD TLS config - client key file path                                                                  |
| --etcd.tls.ca                    | ETCD_TLS_CA                      | False    |                        | ETCD TLS config - CA file path                                                                          |
| --log.json                       | LOG_JSON                         | False    | `false`                | Tell the application to log json, default: false                                                        |
| --log.level                      | LOG_LEVEL                        | False    | `info`                 | The log level to use for filtering logs, possible values: debug, info, warn, error                      |
| --telegram.token                 | TELEGRAM_TOKEN                   | True     |                        | The token used to connect with Telegram. Token you get from [@botfather](https://telegram.me/botfather) |
| --template.path                  | TEMPLATE_PATH                    | True     |                        | The path to the template                                                                                |
| --telegram.admin                 | TELEGRAM_ADMIN                   | True     |                        | Telegram admin IDs                                                                                      |

#### Authentication
Users may be allowed to command the bot specifies by multiply `--telegram.admin` command line option. 

Example:
```bash
grafana-annotations-bot --telegram.admin=123 --telegram.admin=456
```

Or by specifying a newline-separated list of telegram user IDs in the TELEGRAM_ADMIN environment variable.

Example:
```bash
TELEGRAM_ADMIN="123\n456" grafana-annotations-bot
```

#### Message template
Message template specifies by `--template.path` command line option or by TEMPLATE_PATH environment variable.
[Default template](default.tmpl)

##### Template variables
| Go template variable | Type     | Description                                            |
|----------------------|----------|--------------------------------------------------------|
| {{.Title}}           | string   | Annotation title                                       |
| {{.Message}}         | string   | Annotation message                                     |
| {{.Tags}}            | []string | Annotation tags                                        |
| {{.JoinedTags}}      | string   | Annotation tags joined to string by new line separator |
| {{.FormattedDate}}   | string   | Annotation date in RFC1123 format                      |
| {{.Text}}            | string   | Raw annotation body string                             |


## License

  [MIT](LICENSE)