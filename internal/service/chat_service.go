package service

import (
	"context"
	"errors"

	"github.com/azzimoda/rubix-bot/internal/model"
	"github.com/azzimoda/rubix-bot/internal/repository"
	"gorm.io/gorm"
)

// ErrChatNotFound is returned when no chat is registered for a Telegram ID.
var ErrChatNotFound = errors.New("chat not found")

// ChatService manages the lifecycle of chats.
type ChatService struct {
	repo repository.ChatRepository
}

// NewChatService builds a ChatService.
func NewChatService(repo repository.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

// Create registers a chat by its Telegram chat ID.
func (s *ChatService) Create(ctx context.Context, tgChatID int64) (*model.Chat, error) {
	chat := &model.Chat{TgChatID: tgChatID}
	if err := s.repo.Create(ctx, chat); err != nil {
		return nil, err
	}
	return chat, nil
}

// GetByTgChatID returns the chat registered for a Telegram chat ID.
func (s *ChatService) GetByTgChatID(ctx context.Context, tgChatID int64) (*model.Chat, error) {
	chat, err := s.repo.GetByTGID(ctx, tgChatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}
	return chat, nil
}

// Exists reports whether a chat is registered for a Telegram chat ID.
func (s *ChatService) Exists(ctx context.Context, tgChatID int64) (bool, error) {
	_, err := s.GetByTgChatID(ctx, tgChatID)
	if errors.Is(err, ErrChatNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
