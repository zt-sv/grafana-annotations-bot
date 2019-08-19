package telegram

import (
	"context"
	"github.com/13rentgen/grafana-annotations-bot/internal/pkg/database"
	"github.com/13rentgen/grafana-annotations-bot/internal/pkg/grafana"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/tucnak/telebot"
	"html/template"
	"runtime"
	"time"
)

const (
	commandStart  = "/start"
	commandStop   = "/stop"
	commandStatus = "/status"
)

var botVersion = runtime.Version()

// BotOptions : telegram bot config
type BotOptions struct {
	Addr          string
	Token         string
	Store         *database.DbClient
	Logger        log.Logger
	Revision      string
	Template      *template.Template
	GrafanaClient *grafana.Client
	Admins        []int
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
	admins        []int
}

// NewBot : create new telegram bot
func NewBot(options BotOptions) (*Bot, error) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  options.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		return nil, err
	}

	tgBot := &Bot{
		admins:        options.Admins,
		token:         options.Token,
		logger:        options.Logger,
		startTime:     time.Now(),
		template:      options.Template,
		tb:            bot,
		store:         options.Store,
		grafanaClient: options.GrafanaClient,
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

			bot.tb.Handle(commandStart, bot.onlyForAdmins(bot.handleStart))
			bot.tb.Handle(commandStatus, bot.onlyForAdmins(bot.handleStatus))
			bot.tb.Handle(commandStop, bot.onlyForAdmins(bot.handleStop))

			bot.tb.Start()
			return nil
		}, func(err error) {})
	}

	return gr.Run()
}

func (bot *Bot) onlyForAdmins(handler func(m *telebot.Message)) func(m *telebot.Message) {
	return func(m *telebot.Message) {
		if !bot.isAdminID(m.Sender.ID) {
			level.Error(bot.logger).Log("msg", "Receive command from not admin user")
			bot.tb.Send(
				m.Chat,
				"Permission denied",
				&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
			)
			return
		}

		handler(m)
	}
}

// isAdminID returns whether id is one of the configured admin IDs.
func (bot *Bot) isAdminID(id int) bool {
	for _, adminID := range bot.admins {
		if id == adminID {
			return true
		}
	}

	return false
}
