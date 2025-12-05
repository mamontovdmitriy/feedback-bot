package service

import (
	"context"
	"errors"
	"feedback-bot/internal/entity"
	"feedback-bot/internal/repo"
	"sync"
)

type UserThreadService struct {
	repo repo.UserThread
	mu   sync.RWMutex
	list map[int64]int64
}

func NewUserThreadService(repo repo.UserThread) *UserThreadService {
	return &UserThreadService{repo: repo}
}

func (s *UserThreadService) SaveUserThread(userId int64, threadId int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.repo.Add(ctx, entity.UserThread{UserId: userId, ThreadId: threadId})
	if err != nil {
		return err
	}

	s.list[userId] = threadId

	return nil
}

func (s *UserThreadService) GetThreadId(userId int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(s.list) == 0 {
		list, _ := s.repo.GetList(ctx)
		s.list = list
	}

	threadId, exists := s.list[userId]
	if !exists {
		return 0, errors.New("ThreadID does not exist")
	}

	return threadId, nil
}

func (s *UserThreadService) GetUserId(threadId int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(s.list) == 0 {
		list, _ := s.repo.GetList(ctx)
		s.list = list
	}

	// поиск использует меньше итераций, чем разворачивание мапы
	for k, v := range s.list {
		if v == threadId {
			return k, nil
		}
	}

	return 0, errors.New("UserId does not exist")
}
