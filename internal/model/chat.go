package model

import "gorm.io/gorm"

type Chat struct {
	gorm.Model

	// TgChatID holds Telegram chat ID.
	TgChatID int64 `json:"tg_chat_id" gorm:"uniqueIndex"`
}
