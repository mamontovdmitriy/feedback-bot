package handler

import (
	"embed"
	"fmt"
	"html/template"
	"testing"
)

//go:embed templates/*.html
var testhtmlFiles embed.FS

func TestExistTemplates(t *testing.T) {
	templates, _ := template.ParseFS(htmlFiles, "templates/*.html")

	testCases := []string{
		"client-profile.html",
		"cmd-help.html",
		"cmd-info.html",
		"cmd-start.html",
		"cmd-unknown.html",
		"msg-add-bot-to-group.html",
		"msg-error-bad-config.html",
		"msg-error-empty-reply.html",
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Exists template %s", tc), func(t *testing.T) {
			_, err := RenderTemplate(templates, tc, nil)
			if err != nil {
				t.Errorf("RenderTemplate(%s) вернул ошибку %v", tc, err)
			}
		})
	}
}
