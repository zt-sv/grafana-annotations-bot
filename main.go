package main

import (
	"context"
	"fmt"
	"github.com/13rentgen/grafana-annotations-bot/pkg/database"
	"github.com/13rentgen/grafana-annotations-bot/pkg/grafana"
	tg "github.com/13rentgen/grafana-annotations-bot/pkg/telegram"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/joho/godotenv"
	"github.com/oklog/run"
	"gopkg.in/alecthomas/kingpin.v2"
	"html/template"
	"net/url"
	"os"
	"time"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

func main() {
	godotenv.Load()

	config := struct {
		grafanaURL            *url.URL
		grafanaToken          string
		grafanaUseTLS         bool
		grafanaSkipVerify     bool
		grafanaClientKeyFile  string
		grafanaClientCertFile string
		boltPath              string
		logLevel              string
		logJSON               bool
		telegramAdmins        []int
		telegramToken         string
		templatePath          string
		scrapeInterval        time.Duration
	}{}

	a := kingpin.New("grafana-annotations-bot", "Fetch Grafana annotations to telegram")
	a.HelpFlag.Short('h')

	a.Flag("grafana.url", "The URL that's used to connect to the Grafana").
		Required().
		Envar("GRAFANA_URL").
		URLVar(&config.grafanaURL)

	a.Flag("grafana.token", "The Bearer token used to connect with Grafana API").
		Required().
		Envar("GRAFANA_TOKEN").
		StringVar(&config.grafanaToken)

	a.Flag("grafana.scrape_interval", "Scrape annotations interval").
		Envar("SCRAPE_INTERVAL").
		Default("10s").
		DurationVar(&config.scrapeInterval)

	a.Flag("grafana.useTLS", "Use TLS to connect with Grafana").
		Envar("GRAFANA_USE_TLS").
		Default("false").
		BoolVar(&config.grafanaUseTLS)

	a.Flag("grafana.tls_config.insecure_skip_verify", "Grafana TLS config - insecure skip verify").
		Envar("GRAFANA_TLS_INSECURE_SKIP_VERIFY").
		BoolVar(&config.grafanaUseTLS)

	a.Flag("grafana.tls_config.cert_file", "Grafana TLS config - client cert file path").
		Envar("GRAFANA_TLS_CERT_FILE").
		ExistingFileVar(&config.grafanaClientCertFile)

	a.Flag("grafana.tls_config.key_file", "Grafana TLS config - client key file path").
		Envar("GRAFANA_TLS_KEY_FILE").
		ExistingFileVar(&config.grafanaClientKeyFile)

	a.Flag("bolt.path", "The path to the file where bolt persists its data").
		Required().
		Envar("BOLT_PATH").
		StringVar(&config.boltPath)

	a.Flag("log.json", "Tell the application to log json and not key value pairs").
		Envar("LOG_JSON").
		Default("false").
		BoolVar(&config.logJSON)

	a.Flag("log.level", "The log level to use for filtering logs").
		Envar("LOG_LEVEL").
		Default(levelInfo).
		EnumVar(&config.logLevel, levelError, levelWarn, levelInfo, levelDebug)

	a.Flag("telegram.token", "The token used to connect with Telegram").
		Required().
		Envar("TELEGRAM_TOKEN").
		StringVar(&config.telegramToken)

	a.Flag("template.path", "The path to the template").
		Required().
		Envar("TEMPLATE_PATH").
		ExistingFileVar(&config.templatePath)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("error parsing commandline arguments: %v\n", err)
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	levelFilter := map[string]level.Option{
		levelError: level.AllowError(),
		levelWarn:  level.AllowWarn(),
		levelInfo:  level.AllowInfo(),
		levelDebug: level.AllowDebug(),
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	if config.logJSON {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	}

	logger = level.NewFilter(logger, levelFilter[config.logLevel])
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	// Create bolt store
	var boltStore *database.DbClient
	{
		boltStore, err = database.NewDB(
			database.ClientConfig{
				Path:   config.boltPath,
				Logger: log.With(logger, "component", "database"),
			},
		)

		if err != nil {
			level.Error(logger).Log("msg", "failed to create bolt store client", "err", err)
			os.Exit(1)
		}
	}

	// Create grafana client
	var grafanaClient *grafana.Client
	{
		grafanaClient, err = grafana.NewClient(
			grafana.ClientConfig{
				URL:        config.grafanaURL,
				Token:      config.grafanaToken,
				UseTLS:     config.grafanaUseTLS,
				SkipVerify: config.grafanaSkipVerify,
				CertFile:   config.grafanaClientCertFile,
				KeyFile:    config.grafanaClientKeyFile,
				Logger:     log.With(logger, "component", "grafana_client"),
			},
		)

		if err != nil {
			level.Error(logger).Log("msg", "failed to create grafana client", "err", err)
			os.Exit(1)
		}

		grafanaStatus, err := grafanaClient.GetStatus()

		if err != nil {
			level.Error(logger).Log("msg", "failed to get grafana status", "err", err)
			os.Exit(1)
		}

		level.Info(logger).Log(
			"msg", "grafana status",
			"version", grafanaStatus.Version,
			"database", grafanaStatus.Database,
			"commit", grafanaStatus.Commit,
		)
	}

	// Create telegram bot
	var tgBot *tg.Bot
	{
		tpl, err := template.ParseFiles(config.templatePath)

		if err != nil {
			level.Error(logger).Log("msg", "failed to parse template file", "template", config.templatePath, "err", err)
			os.Exit(1)
			return
		}

		tgBot, err = tg.NewBot(
			tg.BotConfig{
				Token:         config.telegramToken,
				Store:         boltStore,
				Logger:        log.With(logger, "component", "telegram_bot"),
				Template:      tpl,
				GrafanaClient: grafanaClient,
			},
		)

		if err != nil {
			level.Error(logger).Log("msg", "failed to start telegram bot", "err", err)
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	annotationsChannel := make(chan grafana.Annotation, 32)

	var gr run.Group

	// Start bot goroutine
	{
		gr.Add(func() error {
			return tgBot.Run(ctx, annotationsChannel)
		}, func(err error) {
			cancel()
		})
	}

	// Scrape grafana goroutine
	{
		ticker := time.NewTicker(config.scrapeInterval)
		lastScrapeTime := time.Now()
		quit := make(chan struct{})

		gr.Add(func() error {
			for {
				select {
				case <-ticker.C:
					currTime := time.Now()
					annotationsResps, err := grafanaClient.GetAnnotations(lastScrapeTime, currTime)

					if err != nil {
						level.Error(logger).Log("msg", "failed to get annotations", "err", err)
					}

					lastScrapeTime = currTime

					for _, annotation := range annotationsResps {
						level.Info(logger).Log("msg", "get new annotation", "annotation", annotation)
						annotationsChannel <- annotation
					}

				case <-quit:
					ticker.Stop()
					return nil
				}
			}
		}, func(err error) {
			cancel()
		})
	}

	// Start
	if err := gr.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
