package main

import (
	"context"
	"fmt"
	"os"
	"time"

	app "github.com/13rentgen/grafana-annotations-bot/internal/app/grafana-annotations-bot"
	"github.com/13rentgen/grafana-annotations-bot/internal/pkg/database"
	"github.com/13rentgen/grafana-annotations-bot/internal/pkg/grafana"
	tg "github.com/13rentgen/grafana-annotations-bot/internal/pkg/telegram"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
)

func main() {
	var gr run.Group
	config, err := app.LoadConfig()

	if err != nil {
		fmt.Printf("error loading configuration: %v\n", err)
		os.Exit(2)
	}

	var logger = app.GetLogger(config)

	// Create BoltDB store
	boltStore, err := database.NewDB(
		database.ClientConfig{
			Path:   config.BoltPath,
			Logger: log.With(logger, "component", "database"),
		},
	)

	if err != nil {
		level.Error(logger).Log("msg", "failed to create bolt store client", "err", err)
		os.Exit(1)
	}

	// Create Grafana client
	grafanaClient, err := grafana.NewClient(
		grafana.ClientConfig{
			URL:        config.GrafanaURL,
			Token:      config.GrafanaToken,
			UseTLS:     config.GrafanaUseTLS,
			SkipVerify: config.GrafanaSkipVerify,
			CertFile:   config.GrafanaClientCertFile,
			KeyFile:    config.GrafanaClientKeyFile,
			Logger:     log.With(logger, "component", "grafana_client"),
		},
	)

	if err != nil {
		level.Error(logger).Log("msg", "failed to create grafana client", "err", err)
		os.Exit(1)
	}

	// Check grafana status
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

	// Create telegram bot
	tgBot, err := tg.NewBot(
		tg.BotOptions{
			Token:         config.TelegramToken,
			Store:         boltStore,
			Logger:        log.With(logger, "component", "telegram_bot"),
			Template:      config.Template,
			GrafanaClient: grafanaClient,
			Admins:        config.TelegramAdmins,
		},
	)

	if err != nil {
		level.Error(logger).Log("msg", "failed to start telegram bot", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	annotationsChannel := make(chan grafana.Annotation, 32)
	ticker := time.NewTicker(config.ScrapeInterval)
	lastScrapeTime := time.Now()
	quit := make(chan struct{})

	// Start bot goroutine
	gr.Add(func() error {
		return tgBot.Run(ctx, annotationsChannel)
	}, func(err error) {
		cancel()
	})

	// Scrape grafana goroutine
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

				for i := len(annotationsResps) - 1; i >= 0; i-- {
					level.Info(logger).Log("msg", "get new annotation", "annotation", annotationsResps[i])
					annotationsChannel <- annotationsResps[i]
				}

			case <-quit:
				ticker.Stop()
				return nil
			}
		}
	}, func(err error) {
		cancel()
	})

	// Start
	if err := gr.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
