package handler

import (
	tg "github.com/OvyFlash/telegram-bot-api"
)

type UnknownCommandHandler struct {
	BaseDependencies
}

func NewUnknownCommandHandler(baseDeps BaseDependencies) *UnknownCommandHandler {
	return &UnknownCommandHandler{BaseDependencies: baseDeps}
}

func (h *UnknownCommandHandler) HandleCommand(message *tg.Message) {
	h.SendTemplate(message.Chat.ID, "cmd-unknown.html", nil)
}
