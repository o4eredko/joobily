package api

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func (h *Handler) Register(i *iris.Application) {
	api := i.Party("/api")
	v1 := api.Party("/v1")

	v1.Get("/ping", h.Ping)
	slack := v1.Party("/slack")

	slack.ConfigureContainer(func(container *router.APIContainer) {
		container.Post("/events", h.SlackEvents)
		container.Post("/dialog", h.SlackDialogs)
		container.Post("/slash/help", h.SlackSlashHelp)
		container.Post("/slash/ask", h.SlackSlashAsk)
		container.Post("/slash/add", h.SlackSlashAdd)
		container.Post("/slash/discard", h.SlackSlashDiscard)
	})
}
