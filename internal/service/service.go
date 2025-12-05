package service

import (
	"feedback-bot/internal/repo"

	"github.com/sirupsen/logrus"
)

type (
	ServicesDependencies struct {
		Repos *repo.Repositories
		// other components
	}

	Services struct {
		Log           *logrus.Logger
		MessageUpdate *MessageUpdateService
		UserThread    *UserThreadService
		// Auth        Auth
	}
)

func NewServices(log *logrus.Logger, deps *ServicesDependencies) *Services {
	return &Services{
		Log:           log,
		MessageUpdate: NewMesssageUpdateService(deps.Repos.MessageUpdate),
		UserThread:    NewUserThreadService(deps.Repos.UserThread),
		// Auth: NewAuthService(deps.Repos.User, deps.Hasher, deps.SignKey, deps.TokenTTL),
	}
}
