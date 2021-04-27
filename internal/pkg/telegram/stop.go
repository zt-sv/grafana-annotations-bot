package telegram

import (
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/tucnak/telebot"
)

func (bot *Bot) handleStop(m *telebot.Message) {
	exist, err := bot.store.ExistChat(m.Chat)

	if err != nil {
		level.Error(bot.logger).Log("msg", "Could check key in store", "err", err)
	}

	if !exist {
		level.Warn(bot.logger).Log("msg", "Chat not subscribed for tags", "chat", m.Chat.ID)
		bot.tb.Send(m.Chat, "You're not subscribed for any tags yet")

		return
	}

	if err != nil {
		level.Error(bot.logger).Log("msg", "Could not remove chat from store", "err", err)
		bot.tb.Send(m.Chat, "Something went wrong...")

		return
	}

	chatTags, err := bot.store.GetChatTags(m.Chat)
	err = bot.store.Remove(m.Chat)

	if chatTags != nil {
		bot.tb.Send(m.Chat, "You're successfully unsubscribe for tags:\n"+strings.Join(chatTags, "\n"))
		return
	}

	bot.tb.Send(m.Chat, "You're successfully unsubscribe")
}
