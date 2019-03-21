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
| Flag                                      | ENV                              | Required | Description                                                                                             |
|-------------------------------------------|----------------------------------|----------|---------------------------------------------------------------------------------------------------------|
| --grafana.url                             | GRAFANA_URL                      | True     | The URL that's used to connect to the Grafana, example: `http://localhost:3000`                         |
| --grafana.token                           | GRAFANA_TOKEN                    | True     | The Bearer token used to connect with Grafana API                                                       |
| --grafana.scrape_interval                 | SCRAPE_INTERVAL                  | False    | Scrape annotations interval. Default: 10s                                                               |
| --grafana.useTLS                          | GRAFANA_USE_TLS                  | False    | Use TLS to connect with Grafana API, default: false                                                     |
| --grafana.tls_config.insecure_skip_verify | GRAFANA_TLS_INSECURE_SKIP_VERIFY | False    | Grafana TLS config - insecure skip verify                                                               |
| --grafana.tls_config.cert_file            | GRAFANA_TLS_CERT_FILE            | False    | Grafana TLS config - client cert file path                                                              |
| --grafana.tls_config.key_file             | GRAFANA_TLS_KEY_FILE             | False    | Grafana TLS config - client key file path                                                               |
| --bolt.path                               | BOLT_PATH                        | True     | Bolt database file path                                                                                 |
| --log.json                                | LOG_JSON                         | False    | Tell the application to log json, default: false                                                        |
| --log.level                               | LOG_LEVEL                        | False    | The log level to use for filtering logs, possible values: debug, info, warn, error. Default: info       |
| --telegram.token                          | TELEGRAM_TOKEN                   | True     | The token used to connect with Telegram. Token you get from [@botfather](https://telegram.me/botfather) |
| --template.path                           | TEMPLATE_PATH                    | True     | The path to the template                                                                                |


