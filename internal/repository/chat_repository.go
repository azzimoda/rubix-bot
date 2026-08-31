package repository

import (
	"context"

	"github.com/azzimoda/rubix-bot/internal/model"
	"gorm.io/gorm"
)

// ChatRepository persists chats.
type ChatRepository interface {
	GetByTGID(ctx context.Context, tgID int64) (*model.Chat, error)
	Create(ctx context.Context, chat *model.Chat) error
}

type gormChatRepository struct {
	db *gorm.DB
}

// NewChatRepository returns a GORM-backed ChatRepository.
func NewChatRepository(db *gorm.DB) ChatRepository {
	return &gormChatRepository{db: db}
}

func (r *gormChatRepository) GetByTGID(ctx context.Context, tgID int64) (*model.Chat, error) {
	var chat model.Chat
	if err := r.db.WithContext(ctx).Where("tg_chat_id = ?", tgID).First(&chat).Error; err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *gormChatRepository) Create(ctx context.Context, chat *model.Chat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}
