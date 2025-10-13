package repository

import (
	"github.com/eyupaydin41/query-service/model"
	"gorm.io/gorm"
)

type LoginHistoryRepository struct {
	db *gorm.DB
}

func NewLoginHistoryRepository(db *gorm.DB) *LoginHistoryRepository {
	err := db.AutoMigrate(&model.LoginHistory{})
	if err != nil {
		return nil
	}
	return &LoginHistoryRepository{db: db}
}

func (r *LoginHistoryRepository) Create(loginHistory *model.LoginHistory) error {
	return r.db.Create(loginHistory).Error
}
