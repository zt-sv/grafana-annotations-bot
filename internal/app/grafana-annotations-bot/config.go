package app

import (
	"fmt"
	"net/url"
	"os"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/joho/godotenv"
)

import (
	"html/template"
)

const (
	// StoreTypeBolt : boltdb store type
	StoreTypeBolt = "bolt"
	// StoreTypeEtcd : etcd store type
	StoreTypeEtcd = "etcd"

	storeKeyPrefix = "annotationsbot/chats"

	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

type boltdbStoreConfig struct {
	Path string
}

type etcdStoreConfig struct {
	URL                   *url.URL
	TLSInsecure           bool
	TLSInsecureSkipVerify bool
	TLSCert               string
	TLSKey                string
	TLSCA                 string
}

type grafanaConfig struct {
	URL                   *url.URL
	Token                 string
	TLSInsecure           bool
	TLSInsecureSkipVerify bool
	TLSCert               string
	TLSKey                string
	ScrapeInterval        time.Duration
}

// StorageConfig : storage configuration
type StorageConfig struct {
	StoreType         string
	StoreKeyPrefix    string
	BoltdbStoreConfig boltdbStoreConfig
	EtcdStoreConfig   etcdStoreConfig
}

// Configuration : main project configuration
type Configuration = struct {
	GrafanaConfig  grafanaConfig
	StorageConfig  StorageConfig
	LogLevel       string
	LogJSON        bool
	TelegramAdmins []int64
	TelegramToken  string
	TemplatePath   string
	Template       *template.Template
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
		URLVar(&config.GrafanaConfig.URL)

	a.Flag("grafana.token", "The Bearer token used to connect with Grafana API").
		Required().
		Envar("GRAFANA_TOKEN").
		StringVar(&config.GrafanaConfig.Token)

	a.Flag("grafana.scrapeInterval", "Scrape annotations interval").
		Envar("GRAFANA_SCRAPE_INTERVAL").
		Default("10s").
		DurationVar(&config.GrafanaConfig.ScrapeInterval)

	a.Flag("grafana.tls.insecure", "Insecure connection to Grafana API").
		Envar("GRAFANA_TLS_INSECURE").
		Default("false").
		BoolVar(&config.GrafanaConfig.TLSInsecure)

	a.Flag("grafana.tls.insecureSkipVerify", "Grafana TLS config - insecure skip verify").
		Default("false").
		Envar("GRAFANA_TLS_INSECURE_SKIP_VERIFY").
		BoolVar(&config.GrafanaConfig.TLSInsecureSkipVerify)

	a.Flag("grafana.tls.cert", "Grafana TLS config - client cert file path").
		Envar("GRAFANA_TLS_CERT").
		ExistingFileVar(&config.GrafanaConfig.TLSCert)

	a.Flag("grafana.tls.key", "Grafana TLS config - client key file path").
		Envar("GRAFANA_TLS_KEY_FILE").
		ExistingFileVar(&config.GrafanaConfig.TLSKey)

	a.Flag("store.type", fmt.Sprintf("The store to use. Possible values %s, %s", StoreTypeBolt, StoreTypeEtcd)).
		Envar("STORE_TYPE").
		Default(StoreTypeBolt).
		EnumVar(&config.StorageConfig.StoreType, StoreTypeBolt, StoreTypeEtcd)

	a.Flag("store.keyPrefix", "Prefix for store keys").
		Envar("STORE_KEY_PREFIX").
		Default(storeKeyPrefix).
		StringVar(&config.StorageConfig.StoreKeyPrefix)

	a.Flag("bolt.path", "The path to the file where bolt persists its data").
		Default("/tmp/bot.db").
		Envar("BOLT_PATH").
		StringVar(&config.StorageConfig.BoltdbStoreConfig.Path)

	a.Flag("etcd.url", "The URL that's used to connect to the etcd store").
		Default("localhost:2379").
		Envar("ETCD_URL").
		URLVar(&config.StorageConfig.EtcdStoreConfig.URL)

	a.Flag("etcd.tls.insecure", "Insecure connection to ETCD").
		Default("false").
		Envar("ETCD_TLS_INSECURE").
		BoolVar(&config.StorageConfig.EtcdStoreConfig.TLSInsecure)

	a.Flag("etcd.tls.insecureSkipVerify", "ETCD TLS config - insecure skip verify").
		Default("false").
		Envar("ETCD_TLS_INSECURE_SKIP_VERIFY").
		BoolVar(&config.StorageConfig.EtcdStoreConfig.TLSInsecureSkipVerify)

	a.Flag("etcd.tls.cert", "ETCD TLS config - client cert file path").
		Envar("ETCD_TLS_CERT").
		StringVar(&config.StorageConfig.EtcdStoreConfig.TLSCert)

	a.Flag("etcd.tls.key", "ETCD TLS config - client key file path").
		Envar("ETCD_TLS_KEY").
		StringVar(&config.StorageConfig.EtcdStoreConfig.TLSKey)

	a.Flag("etcd.tls.ca", "ETCD TLS config - CA file path").
		Envar("ETCD_TLS_CA").
		StringVar(&config.StorageConfig.EtcdStoreConfig.TLSCA)

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
		Int64ListVar(&config.TelegramAdmins)

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
