package handler

import (
	"fmt"

	tg "github.com/OvyFlash/telegram-bot-api"
)

type Criterium struct {
	Title string
	Check func(chm tg.ChatMember) (string, bool)
}

func generateAccessRequirement(chatMember tg.ChatMember, criteria []Criterium) string {
	var msg = ""
	for _, cr := range criteria {
		value, success := cr.Check(chatMember)
		msg += fmt.Sprintf("%s %s: %s\n", getStatusIcon(success), cr.Title, value)
	}
	return msg
}

func getStatusIcon(success bool) string {
	return map[bool]string{true: "✅", false: "❌"}[success]
}
