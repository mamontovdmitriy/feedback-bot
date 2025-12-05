package handler

import (
	tg "github.com/OvyFlash/telegram-bot-api"
)

type GroupChatMemberHandler struct {
	BaseDependencies
}

func NewGroupChatMemberHandler(baseDeps BaseDependencies) *GroupChatMemberHandler {
	return &GroupChatMemberHandler{BaseDependencies: baseDeps}
}

func (h *GroupChatMemberHandler) HandleChatMember(chatMemberUpdated *tg.ChatMemberUpdated) {
	newMember := chatMemberUpdated.NewChatMember

	if !newMember.User.IsBot {
		return
	}

	var criteria = []Criterium{
		Criterium{"Статус", func(chm tg.ChatMember) (string, bool) {
			return chm.Status, chm.IsAdministrator()
		}},
		Criterium{"Управление чатом", func(chm tg.ChatMember) (string, bool) {
			result := "Нет"
			if chm.CanManageChat {
				result = "Да"
			}
			return result, chm.CanManageChat
		}},
		Criterium{"Удаление сообщений", func(chm tg.ChatMember) (string, bool) {
			result := "Нет"
			if chm.CanDeleteMessages {
				result = "Да"
			}
			return result, chm.CanDeleteMessages
		}},
	}

	data := struct {
		AccessRequirement string
		TgChannelId       int64
	}{
		AccessRequirement: generateAccessRequirement(newMember, criteria),
		TgChannelId:       chatMemberUpdated.Chat.ID,
	}

	h.SendTemplate(chatMemberUpdated.Chat.ID, "msg-add-bot-to-group.html", data)
}
