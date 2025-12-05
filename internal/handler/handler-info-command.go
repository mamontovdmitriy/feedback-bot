package handler

import (
	tg "github.com/OvyFlash/telegram-bot-api"
)

type InfoCommandHandler struct {
	BaseDependencies
}

func NewInfoCommandHandler(baseDeps BaseDependencies) *InfoCommandHandler {
	return &InfoCommandHandler{BaseDependencies: baseDeps}
}

func (h *InfoCommandHandler) HandleCommand(message *tg.Message) {
	h.SendTemplate(message.Chat.ID, "cmd-info.html", message.From)
}
