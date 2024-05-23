package telegram

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/zt-sv/grafana-annotations-bot/internal/pkg/grafana"
	"gopkg.in/telebot.v3"
)

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
					bot.tb.Send(
						chatAndTags.Chat,
						renderedTpl,
						&telebot.SendOptions{ParseMode: telebot.ModeHTML, ThreadID: chatAndTags.ThreadId},
					)
				}
			}
		}
	}
}
