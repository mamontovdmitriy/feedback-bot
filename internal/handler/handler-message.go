package handler

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	tg "github.com/OvyFlash/telegram-bot-api"
)

var (
	reExtractUserId = regexp.MustCompile(`(?i)User\s?ID:\s?([-\d]+)`)
)

/**
 * Пересылка сообщения клиента из бота в комментарий соответствующего поста в группе обсуждения
 * Обратная пересылка сообщения менеджера из комментов соответствующему клиенту.
 */
func (c *Handler) proxyMessage(message *tg.Message) {
	// Create forum topic into channel - save thread id
	if message.ForumTopicCreated != nil && message.Chat.ID == c.cfg.SupergroupId {
		c.saveForumThreadId(message)
		return
	}

	// Change forum topic
	if message.ForumTopicEdited != nil || message.ForumTopicClosed != nil || message.ForumTopicReopened != nil {
		c.services.Log.Info("Handler.proxyMessage: ignore changing any forum topics")
		return
	}

	// ignore own messages from Telegram bot
	if message.From == nil || message.From.IsBot {
		c.services.Log.Info("Handler.proxyMessage: ignore messages from Telegram")
		return
	}

	// response to bot
	if message.ReplyToMessage != nil {
		c.copyMessageFromChannelToBot(message)
		return
	}

	// ignore messages in General thread
	if message.Chat.ID == c.cfg.SupergroupId {
		c.SendTemplate(c.cfg.SupergroupId, "msg-error-empty-reply.html", nil)
		c.services.Log.Info("Handler.proxyMessage: ignore messages in General thread")
		return
	}

	// Send message from bot to channel
	userId := message.From.ID
	threadId, err := c.services.UserThread.GetThreadId(userId)
	if err != nil {
		// Create user's forum topic
		name := fmt.Sprintf("%s %s (%s)", message.From.LastName, message.From.FirstName, message.From.UserName)
		_, err := c.CreateForumTopic(userId, name)
		if err != nil {
			c.services.Log.Errorf("Handler.proxyMessage: error creating forum topic - %v", err)
			// todo show resend button
			return
		}

		// Wait for creation of forum topic
		go func() {
			for {
				threadId, err := c.services.UserThread.GetThreadId(userId)
				if err == nil {
					// copy message from Bot into forum topic
					c.copyMessageToThread(threadId, message)
					c.services.Log.Infof("Message async proxied %d", message.MessageID)
					return
				}
				time.Sleep(3 * time.Second)
			}
		}()
	} else {
		// copy message from Bot into forward post comments
		c.copyMessageToThread(threadId, message)
		c.services.Log.Infof("Message proxied %d", message.MessageID)
	}
}

func (c *Handler) saveForumThreadId(message *tg.Message) {
	userId, err := extractUserId(message.ForumTopicCreated.Name)
	if err != nil {
		c.services.Log.Warn("Handler.proxyMessage: userId not found")
		return
	}

	err = c.services.UserThread.SaveUserThread(userId, int64(message.MessageThreadID))
	if err != nil {
		c.services.Log.Errorf("Handler.proxyMessage: error saving forum thread ID %v", err)
	}
}

func (c *Handler) copyMessageFromChannelToBot(message *tg.Message) {
	userId, err := c.services.UserThread.GetUserId(int64(message.ReplyToMessage.MessageThreadID))
	if err != nil {
		c.services.Log.Warn("Handler.proxyMessage: userId not found in ReplyToMessage")

		threadId, err := c.services.UserThread.GetThreadId(message.From.ID)
		if err != nil {
			c.services.Log.Errorf("Handler.proxyMessage: error getting thread id - %v", err)
			// todo show resend button
		}
		c.copyMessageToThread(threadId, message)
		c.services.Log.Infof("Reply message proxied %d", message.MessageID)
		return
	}
	_, err = c.bot.CopyMessage(tg.NewCopyMessage(userId, message.Chat.ID, message.MessageID))
	if err != nil {
		c.services.Log.Errorf("Handler.proxyMessage: error coping reply message to bot - %v", err)
		// todo show resend button
	}
}

func (c *Handler) copyMessageToThread(threadId int64, message *tg.Message) {
	_, err := c.bot.CopyMessage(tg.CopyMessageConfig{
		BaseChat: tg.BaseChat{
			ChatConfig:      tg.ChatConfig{ChatID: c.cfg.SupergroupId},
			ReplyParameters: tg.ReplyParameters{MessageID: int(threadId)},
		},
		FromChat:  tg.ChatConfig{ChatID: message.Chat.ID},
		MessageID: message.MessageID,
	})
	if err != nil {
		c.services.Log.Errorf("Handler.proxyMessage: error coping message to chat - %v", err)

		// show error
		c.SendTemplate(message.Chat.ID, "msg-error-bad-config.html", struct{ Message string }{Message: err.Error()})
	}
}

func extractUserId(text string) (int64, error) {
	match := reExtractUserId.FindStringSubmatch(text)

	if len(match) == 0 {
		return 0, errors.New("regexp failed - userId not found")
	}

	return strconv.ParseInt(match[1], 10, 64)
}
