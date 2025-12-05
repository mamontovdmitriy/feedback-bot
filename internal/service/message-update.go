package service

import (
	"context"
	"feedback-bot/internal/entity"
	"feedback-bot/internal/repo"
)

type MessageUpdateService struct {
	repo repo.MessageUpdate
}

func NewMesssageUpdateService(repo repo.MessageUpdate) *MessageUpdateService {
	return &MessageUpdateService{repo: repo}
}

func (s *MessageUpdateService) Create(entity entity.MessageUpdate) (int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return s.repo.Create(ctx, entity)
}
