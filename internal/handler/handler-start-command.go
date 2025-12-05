package handler

import (
	tg "github.com/OvyFlash/telegram-bot-api"
)

type StartCommandHandler struct {
	BaseDependencies
}

func NewStartCommandHandler(baseDeps BaseDependencies) *StartCommandHandler {
	return &StartCommandHandler{BaseDependencies: baseDeps}
}

func (h *StartCommandHandler) HandleCommand(message *tg.Message) {
	h.SendTemplate(message.Chat.ID, "cmd-start.html", nil)
}
