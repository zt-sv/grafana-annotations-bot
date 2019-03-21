package telegram

import (
	"bytes"
	"context"
	"fmt"
	"github.com/13rentgen/grafana-annotations-bot/pkg/database"
	"github.com/13rentgen/grafana-annotations-bot/pkg/grafana"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/tucnak/telebot"
	"html/template"
	"runtime"
	"strings"
	"time"
)

const (
	commandStart  = "/start"
	commandStop   = "/stop"
	commandStatus = "/status"
)

var botVersion = runtime.Version()

// BotConfig : telegram bot config
type BotConfig struct {
	Addr          string
	Token         string
	Store         *database.DbClient
	Logger        log.Logger
	Revision      string
	Template      *template.Template
	GrafanaClient *grafana.Client
}

// Bot : telegram bot
type Bot struct {
	token         string
	store         *database.DbClient
	logger        log.Logger
	startTime     time.Time
	tb            *telebot.Bot
	template      *template.Template
	grafanaClient *grafana.Client
}

// NewBot : create new telegram bot
func NewBot(config BotConfig) (*Bot, error) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  config.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		return nil, err
	}

	tgBot := &Bot{
		token:         config.Token,
		logger:        config.Logger,
		startTime:     time.Now(),
		template:      config.Template,
		tb:            bot,
		store:         config.Store,
		grafanaClient: config.GrafanaClient,
	}

	return tgBot, nil
}

// Run : start telegram bot
func (bot *Bot) Run(ctx context.Context, annotationsChannel <-chan grafana.Annotation) error {
	var gr run.Group
	{
		gr.Add(func() error {
			return bot.listenAnnotations(ctx, annotationsChannel)
		}, func(err error) {
			level.Error(bot.logger).Log("msg", "listen annotations error", "err", err)
		})
	}
	{
		gr.Add(func() error {
			level.Info(bot.logger).Log("msg", "start telegram bot", "start time", bot.startTime)

			bot.tb.Handle(commandStart, bot.handleStart)
			bot.tb.Handle(commandStatus, bot.handleStatus)
			bot.tb.Handle(commandStop, bot.handleStop)

			bot.tb.Start()
			return nil
		}, func(err error) {})
	}

	return gr.Run()
}

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
			"*Grafana*\nVersion: %s\nDatabase: %s\n\n*Telegram Bot*\nVersion: %s\nUptime: %s",
			grafanaStatus.Version,
			grafanaStatus.Database,
			botVersion,
			bot.startTime.Format(time.RFC1123),
		),
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
	)
}

type templateData struct {
	Text      string
	Tags      []string
	Timestamp int64
}

// Title : get formatted title for template
func (t *templateData) Title() string {
	return strings.Split(t.Text, "\n")[0]
}

// Message : get formatted message body for template
func (t *templateData) Message() string {
	return strings.Join(strings.Split(t.Text, "\n")[1:], "\n")
}

// JoinedTags : get formatted tags list for template
func (t *templateData) JoinedTags() string {
	return strings.Join(t.Tags, "\n")
}

// FormattedDate : get formatted date for template
func (t *templateData) FormattedDate() string {
	date := time.Unix(0, t.Timestamp*int64(time.Millisecond))
	return date.Format(time.RFC1123)
}

func tagInList(tag string, tagList []string) bool {
	for _, t := range tagList {
		if t == tag {
			return true
		}
	}

	return false
}

func allTagsExist(annotationsTags []string, subscribedTags []string) bool {
	for _, tag := range subscribedTags {
		if !tagInList(tag, annotationsTags) {
			return false
		}
	}

	return true
}

func (bot *Bot) listenAnnotations(ctx context.Context, annotationsChannel <-chan grafana.Annotation) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case annotation := <-annotationsChannel:
			var tpl bytes.Buffer
			bot.template.Execute(&tpl, &templateData{
				Text:      annotation.Text,
				Tags:      annotation.Tags,
				Timestamp: annotation.Time,
			})
			renderedTpl := tpl.String()

			chatAndTagsList, err := bot.store.List()

			if err != nil {
				level.Error(bot.logger).Log("msg", "failed to get list of chats", "err", err)
			}

			for _, chatAndTags := range chatAndTagsList {
				if allTagsExist(annotation.Tags, chatAndTags.Tags) {
					bot.tb.Send(chatAndTags.Chat, renderedTpl, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
				}
			}
		}
	}
}
