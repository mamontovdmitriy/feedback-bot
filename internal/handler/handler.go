package handler

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"runtime/debug"

	"feedback-bot/config"
	"feedback-bot/internal/entity"
	"feedback-bot/internal/service"

	tg "github.com/OvyFlash/telegram-bot-api"
)

type (
	CallbackHandler interface {
		HandleCallback(callback *tg.CallbackQuery)
	}
	ChatMemberHandler interface {
		HandleChatMember(chatMember *tg.ChatMemberUpdated)
	}
	CommandHandler interface {
		HandleCommand(message *tg.Message)
	}
	MessageHandler interface {
		CommandHandler
	}

	UpdateHandler interface {
		CallbackHandler
		ChatMemberHandler
		MessageHandler
	}

	BaseDependencies struct {
		cfg       config.TG
		bot       *tg.BotAPI
		services  *service.Services
		templates *template.Template
	}

	Handler struct {
		BaseDependencies

		defaultUpdateHandler UpdateHandler
		// Commands
		startCommandHandler   CommandHandler
		helpCommandHandler    CommandHandler
		infoCommandHandler    CommandHandler
		unknownCommandHandler CommandHandler
		// Chat member add/edit
		groupChatMemberHandler ChatMemberHandler
	}
)

//go:embed templates/*.html
var htmlFiles embed.FS

func NewHandler(cfg config.TG, bot *tg.BotAPI, services *service.Services) *Handler {
	templates, err := template.ParseFS(htmlFiles, "templates/*.html")
	if err != nil {
		services.Log.Fatal("Handler.NewHandler: error parsing templates - ", err)
	}

	baseDeps := BaseDependencies{
		cfg:       cfg,
		bot:       bot,
		services:  services,
		templates: templates,
	}

	return &Handler{
		BaseDependencies: baseDeps,

		defaultUpdateHandler: nil,
		// Commands
		startCommandHandler:   NewStartCommandHandler(baseDeps),
		helpCommandHandler:    NewHelpCommandHandler(baseDeps),
		infoCommandHandler:    NewInfoCommandHandler(baseDeps),
		unknownCommandHandler: NewUnknownCommandHandler(baseDeps),
		// Chat member add/edit
		groupChatMemberHandler: NewGroupChatMemberHandler(baseDeps),
	}
}

func (c *Handler) Handle(update tg.Update) {
	defer func() {
		if panicValue := recover(); panicValue != nil {
			c.services.Log.Fatalf("recovered from panic: %v\n%v", panicValue, string(debug.Stack()))
		}
	}()

	// Save all request in db
	c.saveMessageUpdate(update)

	// Processing message or callback
	switch {
	case update.Message != nil:
		c.handleMessage(update.Message)
	case update.EditedMessage != nil:
		c.handleMessage(update.EditedMessage)
	case update.MyChatMember != nil:
		c.handleChatMember(update.MyChatMember)
	}
	// todo update status of message
}

func (c *Handler) saveMessageUpdate(update tg.Update) {
	content, err := json.Marshal(update)
	if err != nil {
		c.services.Log.Warnf("can not marshal to json: %v", update)
		return
	}
	c.services.MessageUpdate.Create(entity.MessageUpdate{
		Id:      update.UpdateID,
		Message: string(content),
	})
}

func (c *Handler) handleMessage(message *tg.Message) {
	if !message.IsCommand() {
		c.proxyMessage(message)
		return
	}

	switch message.Command() {
	case "start":
		c.startCommandHandler.HandleCommand(message)
	case "help":
		c.helpCommandHandler.HandleCommand(message)
	case "info":
		c.infoCommandHandler.HandleCommand(message)
	default:
		c.services.Log.Warnf("Handler.handleCommand: unknown command - %s", message.Command())
		c.unknownCommandHandler.HandleCommand(message)
	}
}

func (c *Handler) handleChatMember(chatMember *tg.ChatMemberUpdated) {
	switch chatMember.Chat.Type {
	case "supergroup": // forum group for post`s comments
		c.groupChatMemberHandler.HandleChatMember(chatMember)
	}
}

func (bh *BaseDependencies) CreateForumTopic(userId int64, name string) (tg.ForumTopic, error) {
	text, err := RenderTemplate(bh.templates, "client-profile.html", struct {
		UserID int64
		Name   string
	}{UserID: userId, Name: name})
	if err != nil {
		bh.services.Log.Error("client-profile.html: error template - ", err)
	}

	return bh.bot.CreateForumTopic(tg.CreateForumTopicConfig{
		ChatConfig: tg.ChatConfig{ChatID: bh.cfg.SupergroupId},
		Name:       text,
	})
}

func (bh *BaseDependencies) SendTemplate(chatID int64, tmplName string, data interface{}) {
	text, err := RenderTemplate(bh.templates, tmplName, data)
	if err != nil {
		bh.services.Log.Error(tmplName, ": error template - ", err)
	}

	_, err = bh.bot.Send(tg.MessageConfig{
		BaseChat: tg.BaseChat{
			ChatConfig: tg.ChatConfig{ChatID: chatID},
			// ReplyParameters: tg.ReplyParameters{MessageID: messageId},
		},
		Text:      text,
		ParseMode: "HTML",
		LinkPreviewOptions: tg.LinkPreviewOptions{
			IsDisabled: false,
		},
	})
	if err != nil {
		bh.services.Log.Error(tmplName, ": error sending - ", err)
		bh.bot.Send(tg.NewMessage(chatID, fmt.Sprintf("❗ERROR: %v", err)))
	}
}

func RenderTemplate(templates *template.Template, templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
