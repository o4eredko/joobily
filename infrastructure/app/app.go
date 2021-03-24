package app

import (
	questionRepo "gitlab.jooble.com/marketing_tech/joobily/adapter/repository/question"
	"gitlab.jooble.com/marketing_tech/joobily/config"
	"gitlab.jooble.com/marketing_tech/joobily/infrastructure/log"
	"gitlab.jooble.com/marketing_tech/joobily/infrastructure/slack"
	"gitlab.jooble.com/marketing_tech/joobily/infrastructure/store/elastic"
	questionUseCase "gitlab.jooble.com/marketing_tech/joobily/usecase/question"
)

type (
	shutdowner interface {
		Shutdown()
	}

	App struct {
		config          *config.Config
		cleanupTasks    []shutdowner
		QuestionUseCase questionUseCase.UseCase
		Slack           *slack.Slack
	}
)

func New(config *config.Config) *App {
	log.ConfigureLogger(config.Logger.Level)

	app := &App{config: config}
	defer app.shutdownOnPanic()
	elasticStore := elastic.New(config.Elastic.DSN())
	app.AddCleanupTask(elasticStore)

	questionRepo := questionRepo.New(elasticStore.Client)

	app.QuestionUseCase = questionUseCase.New(questionRepo)
	app.Slack = slack.New(config)

	return app
}

func (a *App) AddCleanupTask(s shutdowner) {
	a.cleanupTasks = append(a.cleanupTasks, s)
}

func (a *App) Shutdown() {
	lastIndex := len(a.cleanupTasks) - 1

	for i := range a.cleanupTasks {
		a.cleanupTasks[lastIndex-i].Shutdown()
	}
}

func (a *App) shutdownOnPanic() {
	if r := recover(); r != nil {
		a.Shutdown()
		panic(r)
	}
}
