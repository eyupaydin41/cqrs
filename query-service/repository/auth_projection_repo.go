package repository

import (
	"fmt"

	"github.com/eyupaydin41/query-service/model"
	"gorm.io/gorm"
)

type AuthProjectionRepository struct {
	db *gorm.DB
}

func NewAuthProjectionRepository(db *gorm.DB) *AuthProjectionRepository {
	return &AuthProjectionRepository{db: db}
}

// CreateTable - Auth projection tablosunu oluşturur
func (r *AuthProjectionRepository) CreateTable() error {
	return r.db.AutoMigrate(&model.AuthProjection{})
}

// Upsert - Auth projection'ı oluşturur veya günceller
func (r *AuthProjectionRepository) Upsert(auth *model.AuthProjection) error {
	return r.db.Save(auth).Error
}

// UpdateEmail - Email'i günceller
func (r *AuthProjectionRepository) UpdateEmail(id, email string) error {
	result := r.db.Model(&model.AuthProjection{}).Where("id = ?", id).Updates(map[string]interface{}{
		"email":      email,
		"updated_at": gorm.Expr("NOW()"),
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("auth projection not found: %s", id)
	}

	return nil
}

// UpdatePassword - Password hash'i günceller
func (r *AuthProjectionRepository) UpdatePassword(id, passwordHash string) error {
	result := r.db.Model(&model.AuthProjection{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash": passwordHash,
		"updated_at":    gorm.Expr("NOW()"),
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("auth projection not found: %s", id)
	}

	return nil
}

// UpdateStatus - Status'u günceller
func (r *AuthProjectionRepository) UpdateStatus(id, status string) error {
	result := r.db.Model(&model.AuthProjection{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": gorm.Expr("NOW()"),
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("auth projection not found: %s", id)
	}

	return nil
}

// FindByEmail - Email'e göre auth projection'ı bulur
func (r *AuthProjectionRepository) FindByEmail(email string) (*model.AuthProjection, error) {
	var auth model.AuthProjection
	result := r.db.Where("email = ?", email).First(&auth)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found with email: %s", email)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &auth, nil
}

// FindByID - ID'ye göre auth projection'ı bulur
func (r *AuthProjectionRepository) FindByID(id string) (*model.AuthProjection, error) {
	var auth model.AuthProjection
	result := r.db.Where("id = ?", id).First(&auth)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found with id: %s", id)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &auth, nil
}

// Delete - Auth projection'ı siler (soft delete için status güncelleme kullan)
func (r *AuthProjectionRepository) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&model.AuthProjection{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("auth projection not found: %s", id)
	}

	return nil
}
