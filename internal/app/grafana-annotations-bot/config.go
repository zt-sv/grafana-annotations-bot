package app

import (
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/alecthomas/kingpin.v2"
)

import (
	"html/template"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

// Configuration : main project configuration
type Configuration = struct {
	GrafanaURL            *url.URL
	GrafanaToken          string
	GrafanaUseTLS         bool
	GrafanaSkipVerify     bool
	GrafanaClientKeyFile  string
	GrafanaClientCertFile string
	BoltPath              string
	LogLevel              string
	LogJSON               bool
	TelegramAdmins        []int
	TelegramToken         string
	TemplatePath          string
	ScrapeInterval        time.Duration
	Template              *template.Template
}

// LoadConfig : load application config
func LoadConfig() (Configuration, error) {
	var config = Configuration{}
	godotenv.Load()

	a := kingpin.New("grafana-annotations-bot", "Fetch Grafana annotations to telegram")
	a.HelpFlag.Short('h')

	a.Flag("grafana.url", "The URL that's used to connect to the Grafana").
		Required().
		Envar("GRAFANA_URL").
		URLVar(&config.GrafanaURL)

	a.Flag("grafana.token", "The Bearer token used to connect with Grafana API").
		Required().
		Envar("GRAFANA_TOKEN").
		StringVar(&config.GrafanaToken)

	a.Flag("grafana.scrape_interval", "Scrape annotations interval").
		Envar("SCRAPE_INTERVAL").
		Default("10s").
		DurationVar(&config.ScrapeInterval)

	a.Flag("grafana.useTLS", "Use TLS to connect with Grafana").
		Envar("GRAFANA_USE_TLS").
		Default("false").
		BoolVar(&config.GrafanaUseTLS)

	a.Flag("grafana.tls_config.insecure_skip_verify", "Grafana TLS config - insecure skip verify").
		Envar("GRAFANA_TLS_INSECURE_SKIP_VERIFY").
		BoolVar(&config.GrafanaUseTLS)

	a.Flag("grafana.tls_config.cert_file", "Grafana TLS config - client cert file path").
		Envar("GRAFANA_TLS_CERT_FILE").
		ExistingFileVar(&config.GrafanaClientCertFile)

	a.Flag("grafana.tls_config.key_file", "Grafana TLS config - client key file path").
		Envar("GRAFANA_TLS_KEY_FILE").
		ExistingFileVar(&config.GrafanaClientKeyFile)

	a.Flag("bolt.path", "The path to the file where bolt persists its data").
		Required().
		Envar("BOLT_PATH").
		StringVar(&config.BoltPath)

	a.Flag("log.json", "Tell the application to log json and not key value pairs").
		Envar("LOG_JSON").
		Default("false").
		BoolVar(&config.LogJSON)

	a.Flag("log.level", "The log level to use for filtering logs").
		Envar("LOG_LEVEL").
		Default(levelInfo).
		EnumVar(&config.LogLevel, levelError, levelWarn, levelInfo, levelDebug)

	a.Flag("telegram.token", "The token used to connect with Telegram").
		Required().
		Envar("TELEGRAM_TOKEN").
		StringVar(&config.TelegramToken)

	a.Flag("template.path", "The path to the template").
		Required().
		Envar("TEMPLATE_PATH").
		ExistingFileVar(&config.TemplatePath)

	a.Flag("telegram.admin", "The Telegram Admin ID").
		Required().
		Envar("TELEGRAM_ADMIN").
		IntsVar(&config.TelegramAdmins)

	_, err := a.Parse(os.Args[1:])

	if err != nil {
		return config, err
	}

	// Check template
	tpl, err := template.ParseFiles(config.TemplatePath)

	if err != nil {
		return config, err
	}

	config.Template = tpl

	return config, err
}
