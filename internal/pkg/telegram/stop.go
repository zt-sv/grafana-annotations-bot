package telegram

import (
	"strings"

	"github.com/go-kit/kit/log/level"
	"gopkg.in/telebot.v3"
)

func (bot *Bot) handleStop(m *telebot.Message) error {
	exist, err := bot.store.ExistChat(m.Chat)

	if err != nil {
		level.Error(bot.logger).Log("msg", "Could check key in store", "err", err)
	}

	if !exist {
		level.Warn(bot.logger).Log("msg", "Chat not subscribed for tags", "chat", m.Chat.ID)
		_, err := bot.tb.Send(m.Chat, "You're not subscribed for any tags yet")
		return err
	}

	if err != nil {
		level.Error(bot.logger).Log("msg", "Could not remove chat from store", "err", err)
		_, err := bot.tb.Send(m.Chat, "Something went wrong...")
		return err
	}

	chatTags, err := bot.store.GetChatTags(m.Chat)
	err = bot.store.Remove(m.Chat)

	if chatTags != nil {
		_, err := bot.tb.Send(m.Chat, "You're successfully unsubscribe for tags:\n"+strings.Join(chatTags, "\n"))
		return err
	}

	_, err = bot.tb.Send(m.Chat, "You're successfully unsubscribe")
	return err
}
