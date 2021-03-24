package slack

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"gitlab.jooble.com/marketing_tech/joobily/config"
)

type Slack struct {
	*slack.Client
	VerificationToken string
	BotID             string
	SigningSecret     string
}

func New(config *config.Config) *Slack {
	return &Slack{
		Client: slack.New(
			config.Slack.AccessToken,
			slack.OptionDebug(config.Slack.Debug),
		),
		VerificationToken: config.Slack.VerificationToken,
		BotID:             config.Slack.BotID,
		SigningSecret:     config.Slack.SigningSecret,
	}
}

func (s *Slack) VerifySecret(r *http.Request) error {
	sv, err := slack.NewSecretsVerifier(r.Header, s.SigningSecret)
	if err != nil {
		return err
	}

	reader := io.TeeReader(r.Body, &sv)
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err := sv.Ensure(); err != nil {
		return err
	}

	return nil
}

func (s *Slack) ShouldSkipMessage(e *slackevents.MessageEvent) bool {
	return e.User == "" ||
		e.BotID == s.BotID ||
		e.Text == "" ||
		strings.HasPrefix(e.Text, "/")
}

func (s *Slack) IsRetrying(r *http.Request) bool {
	_, ok := r.Header["X-Slack-Retry-Num"]
	return ok
}

func (s *Slack) ParseSlash(r *http.Request) (*slack.SlashCommand, error) {
	if err := s.VerifySecret(r); err != nil {
		return nil, err
	}

	slash, err := slack.SlashCommandParse(r)
	if err != nil {
		return nil, err
	}

	return &slash, nil
}

func (s *Slack) ParseEvent(r *http.Request) (*slackevents.EventsAPIEvent, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(buf)

	event, err := slackevents.ParseEvent(
		buf.Bytes(),
		slackevents.OptionVerifyToken(
			&slackevents.TokenComparator{
				VerificationToken: s.VerificationToken,
			},
		),
	)
	return &event, err
}

func (s *Slack) ParseDialog(r *http.Request) (*slack.InteractionCallback, error) {
	if err := s.VerifySecret(r); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Parse request body
	str, _ := url.QueryUnescape(string(body))
	str = strings.TrimPrefix(str, "payload=")
	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(str), &message); err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *Slack) GetCreateForm() *slack.Dialog {
	textInput := slack.NewTextInput("question", "Question", "")
	textareaInput := slack.NewTextAreaInput("answer", "Answer", "")
	elements := []slack.DialogElement{
		textInput,
		textareaInput,
	}
	return &slack.Dialog{
		CallbackID:  "Callback_ID",
		Title:       "Create Question",
		SubmitLabel: "Submit",
		Elements:    elements,
	}
}
