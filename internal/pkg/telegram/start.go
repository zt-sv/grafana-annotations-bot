package telegram

import (
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/tucnak/telebot"
)

func (bot *Bot) handleStart(m *telebot.Message) {
	if m.Payload == "" {
		level.Warn(bot.logger).Log("msg", "Chat not provide tags for subscribe", "chat", m.Chat.ID)
		bot.tb.Send(
			m.Chat,
			"*You're not provide any tag*\nPlease, provide tags.\n\n*Example:*\n/start tagName,anotherOneTag",
			&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
		)

		return
	}

	exist, err := bot.store.ExistChat(m.Chat)

	if err != nil {
		level.Error(bot.logger).Log("msg", "Could check key in store", "err", err)
	}

	if exist {
		level.Warn(bot.logger).Log("msg", "Chat already subscribed for tags", "chat", m.Chat.ID)
		chatTags, err := bot.store.GetChatTags(m.Chat)

		if err != nil {
			bot.tb.Send(m.Chat, "You're already subscribe for tags\n\nUnsubscribe first")
			return
		}

		bot.tb.Send(m.Chat, "You're already subscribe for tags:\n"+strings.Join(chatTags, "\n")+"\n\nUnsubscribe first")

		return
	}

	tags := strings.Split(m.Payload, ",")
	bot.store.AddChatTags(m.Chat, tags)
	bot.tb.Send(m.Chat, "You're subscribe for tags:\n"+strings.Join(tags, "\n"))
}
