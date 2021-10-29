package telegram

import (
	"fmt"
	"time"

	"github.com/13rentgen/grafana-annotations-bot/internal/pkg/build"
	"github.com/go-kit/kit/log/level"
	"github.com/tucnak/telebot"
)

func (bot *Bot) handleStatus(m *telebot.Message) {
	grafanaStatus, err := bot.grafanaClient.GetStatus()

	if err != nil {
		level.Warn(bot.logger).Log("msg", "failed to get grafana status", "err", err)
		bot.tb.Send(m.Chat, fmt.Sprintf("failed to get status... %v", err), nil)
		return
	}

	bot.tb.Send(
		m.Chat,
		fmt.Sprintf(
			"*Grafana*\nVersion: %s\nDatabase: %s\n\n*Telegram Bot*\nVersion: %s\nBuild date: %s\nGo version: %s\nUptime: %s",
			grafanaStatus.Version,
			grafanaStatus.Database,
			build.Version,
			build.BuildDate,
			build.GoVersion,
			bot.startTime.Format(time.RFC1123),
		),
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
	)
}
