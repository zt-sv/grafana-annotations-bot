package telegram

import (
	"fmt"
	"strings"

	"github.com/go-kit/kit/log/level"
	"gopkg.in/telebot.v3"
)

func (bot *Bot) handleStart(m *telebot.Message) error {
	if m.Payload == "" {
		level.Warn(bot.logger).Log("msg", "Chat not provide tags for subscribe", "chat", m.Chat.ID)
		_, err := bot.tb.Send(
			m.Chat,
			"*You're not provide any tag*\nPlease, provide tags.\n\n*Example:*\n/start tagName,anotherOneTag",
			&telebot.SendOptions{ParseMode: telebot.ModeMarkdown, ThreadID: m.ThreadID},
		)
		return err
	}

	exist, err := bot.store.ExistChat(m.Chat, m.ThreadID)

	if err != nil {
		level.Error(bot.logger).Log("msg", "Could check key in store", "err", err)
		return err
	}

	if exist {
		level.Warn(bot.logger).Log("msg", "Chat already subscribed for tags", "chat", m.Chat.ID)
		chatTags, err := bot.store.GetChatTags(m.Chat, m.ThreadID)

		if err != nil {
			bot.tb.Send(
				m.Chat,
				"You're already subscribed for tags\n\nUnsubscribe first",
				&telebot.SendOptions{ParseMode: telebot.ModeMarkdown, ThreadID: m.ThreadID},
			)
			return err
		}

		_, err = bot.tb.Send(
			m.Chat,
			fmt.Sprintf("You're already subscribed for tags:\n%v.\n\nUnsubscribe first", chatTags),
			&telebot.SendOptions{ThreadID: m.ThreadID},
		)

		return err
	}

	tags := strings.Split(m.Payload, ",")
	bot.store.AddChatTags(m.Chat, m.ThreadID, tags)
	_, err = bot.tb.Send(
		m.Chat,
		fmt.Sprintf("You're successfully subscribed for tags:\n%v", tags),
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown, ThreadID: m.ThreadID},
	)
	return err
}
