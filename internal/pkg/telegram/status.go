package telegram

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/zt-sv/grafana-annotations-bot/internal/pkg/build"
	telebot "gopkg.in/telebot.v3"
)

func (bot *Bot) handleStatus(m *telebot.Message) error {
	grafanaStatus, err := bot.grafanaClient.GetStatus()

	if err != nil {
		level.Warn(bot.logger).Log("msg", "failed to get grafana status", "err", err)
		_, err2 := bot.tb.Send(
			m.Chat,
			fmt.Sprintf("failed to get status... %v", err),
			&telebot.SendOptions{ParseMode: telebot.ModeMarkdown, ThreadID: m.ThreadID},
		)
		if err2 != nil {
			return err2
		}
		return err
	}

	_, err = bot.tb.Send(
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
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown, ThreadID: m.ThreadID},
	)

	return err
}
