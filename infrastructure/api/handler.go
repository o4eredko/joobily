package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"gitlab.jooble.com/marketing_tech/joobily/domain"
	"gitlab.jooble.com/marketing_tech/joobily/domain/reply"
	slackClient "gitlab.jooble.com/marketing_tech/joobily/infrastructure/slack"
	"gitlab.jooble.com/marketing_tech/joobily/usecase/question"

	"gitlab.jooble.com/marketing_tech/joobily/domain/entity"
)

type Handler struct {
	client          *slackClient.Slack
	questionUseCase question.UseCase
}

func NewHandler(slackClient *slackClient.Slack, questionUseCase question.UseCase) *Handler {
	return &Handler{
		client:          slackClient,
		questionUseCase: questionUseCase,
	}
}

func (h *Handler) Ping(c iris.Context) {
	c.JSON(iris.Map{"pong": true})
}

func (h *Handler) SlackSlashAsk(c iris.Context) {
	slash, err := h.client.ParseSlash(c.Request())
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	reply, err := h.questionUseCase.Get(slash.Text)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	params := &slack.Msg{
		Type:         slack.MarkdownType,
		Channel:      slash.ChannelID,
		Text:         reply,
		ResponseType: slack.ResponseTypeInChannel,
	}
	c.JSON(params)
}

func (h *Handler) SlackSlashAdd(c iris.Context) {
	slash, err := h.client.ParseSlash(c.Request())
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	reply, err := h.questionUseCase.AddQuestion(slash.UserID, slash.Text)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	params := &slack.Msg{
		Type:         slack.MarkdownType,
		Channel:      slash.ChannelID,
		Text:         reply,
		ResponseType: slack.ResponseTypeInChannel,
	}
	c.JSON(params)
}

func (h *Handler) SlackSlashDiscard(c iris.Context) {
	slash, err := h.client.ParseSlash(c.Request())
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	reply, err := h.questionUseCase.DiscardQuestion(slash.UserID)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	params := &slack.Msg{
		Type:         slack.MarkdownType,
		Channel:      slash.ChannelID,
		Text:         reply,
		ResponseType: slack.ResponseTypeInChannel,
	}
	c.JSON(params)
}

func (h *Handler) SlackSlashHelp(c iris.Context) {
	slash, err := h.client.ParseSlash(c.Request())
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	params := &slack.Msg{
		Type:         slack.MarkdownType,
		Channel:      slash.ChannelID,
		Text:         reply.Help,
		ResponseType: slack.ResponseTypeInChannel,
	}
	c.JSON(params)
}

func (h *Handler) SlackDialogs(c iris.Context) {
	message, err := h.client.ParseDialog(c.Request())
	if err != nil {
		return
	}

	switch message.Type {
	case slack.InteractionTypeShortcut:
		// Make new dialog components and open a dialog.
		dialog := h.client.GetCreateForm()
		h.client.OpenDialog(message.TriggerID, *dialog)
	case slack.InteractionTypeDialogSubmission:
		var input entity.Question
		if err := mapstructure.Decode(&message.Submission, &input); err != nil {
			log.Error().Msg(err.Error())
		}
		input.Question = strings.ToLower(input.Question)
		if _, err := h.questionUseCase.Upsert(&input); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func (h *Handler) SlackEvents(c iris.Context) {
	request := c.Request()
	if h.client.IsRetrying(request) {
		log.Info().Msg("slack retry, skipping")
		return
	}
	eventsAPIEvent, err := h.client.ParseEvent(request)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		challenge := new(slackevents.ChallengeResponse)
		body, _ := ioutil.ReadAll(request.Body)
		json.Unmarshal(body, challenge)
		c.Text(challenge.Challenge, http.StatusOK)
		return
	case slackevents.CallbackEvent:
		switch ev := eventsAPIEvent.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			var msg string
			if h.client.ShouldSkipMessage(ev) {
				return
			}
			if _, err := h.questionUseCase.GetQuestion(ev.User); err != nil {
				if err != domain.EntityNotFound {
					log.Error().Msg(err.Error())
					return
				}
				msg, err = h.questionUseCase.Get(ev.Text)
				if err != nil {
					log.Error().Msg(err.Error())
				}
			} else {
				msg, err = h.questionUseCase.AddAnswer(ev.User, ev.Text)
				if err != nil {
					log.Error().Msg(err.Error())
					return
				}
			}

			_, _, err = h.client.PostMessage(ev.Channel, slack.MsgOptionText(msg, false))
			if err != nil {
				log.Error().Msg(err.Error())
			}
		}
	}

	c.StatusCode(http.StatusOK)
}
