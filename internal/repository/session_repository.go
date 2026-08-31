package repository

import (
	"context"

	"github.com/azzimoda/rubix-bot/internal/model"
	"gorm.io/gorm"
)

// SessionRepository persists game sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id uint, out *model.Session) error
	GetActiveByChat(ctx context.Context, chatID uint) (*model.Session, error)
	Update(ctx context.Context, session *model.Session) error
	// TrimHistory deletes a chat's ended sessions beyond the newest keep,
	// leaving the active session untouched.
	TrimHistory(ctx context.Context, chatID uint, keep int) error
}

type gormSessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository returns a GORM-backed SessionRepository.
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &gormSessionRepository{db: db}
}

func (r *gormSessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *gormSessionRepository) GetByID(ctx context.Context, id uint, out *model.Session) error {
	return r.db.WithContext(ctx).First(out, id).Error
}

func (r *gormSessionRepository) GetActiveByChat(ctx context.Context, chatID uint) (*model.Session, error) {
	var session model.Session
	if err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Where("ended = ?", false).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormSessionRepository) Update(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *gormSessionRepository) TrimHistory(ctx context.Context, chatID uint, keep int) error {
	var ids []uint
	if err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("chat_id = ? AND ended = ?", chatID, true).
		Order("created_at desc").
		Offset(keep).
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&model.Session{}).Error
}
