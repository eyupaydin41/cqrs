package domain

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserAggregate - User'ın domain logic'ini içeren aggregate root
type UserAggregate struct {
	ID           string
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      uint32

	// Henüz persist edilmemiş event'lar
	uncommittedChanges []DomainEvent
}

// NewUserAggregate - Yeni bir aggregate oluşturur
func NewUserAggregate(id string) *UserAggregate {
	return &UserAggregate{
		ID:                 id,
		Status:             "new",
		uncommittedChanges: []DomainEvent{},
	}
}

// GetUncommittedChanges - Henüz kaydedilmemiş event'ları döner
func (u *UserAggregate) GetUncommittedChanges() []DomainEvent {
	return u.uncommittedChanges
}

// MarkChangesAsCommitted - Event'lar persist edildikten sonra temizler
func (u *UserAggregate) MarkChangesAsCommitted() {
	u.uncommittedChanges = []DomainEvent{}
}

// LoadFromHistory - Event history'den aggregate'i yeniden oluşturur
func (u *UserAggregate) LoadFromHistory(events []DomainEvent) {
	for _, event := range events {
		u.applyChange(event, false)
	}
}

// Register - Yeni user kaydı (Command handler'dan çağrılır)
func (u *UserAggregate) Register(email, password string) error {
	// Business rule: User zaten kayıtlı olmamalı
	if u.Status != "new" {
		return errors.New("user already exists")
	}

	// Business rule: Email boş olmamalı
	if email == "" {
		return errors.New("email cannot be empty")
	}

	// Business rule: Password minimum 6 karakter olmalı
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	// Password'ü hash'le
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Event oluştur ve uygula
	event := UserCreatedEvent{
		BaseEvent: BaseEvent{
			AggregateID: u.ID,
			Timestamp:   time.Now(),
			Version:     u.Version + 1,
		},
		Email:        email,
		PasswordHash: string(hash),
	}

	u.applyChange(event, true)
	return nil
}

// ChangePassword - Şifre değiştirme
func (u *UserAggregate) ChangePassword(oldPassword, newPassword string) error {
	// Business rule: User aktif olmalı
	if u.Status != "active" {
		return errors.New("user is not active")
	}

	// Business rule: Eski şifre doğru olmalı
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("invalid old password")
	}

	// Business rule: Yeni şifre minimum 6 karakter olmalı
	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
	}

	// Business rule: Yeni şifre eski şifreden farklı olmalı
	if oldPassword == newPassword {
		return errors.New("new password must be different from old password")
	}

	// Yeni şifreyi hash'le
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Event oluştur ve uygula
	event := PasswordChangedEvent{
		BaseEvent: BaseEvent{
			AggregateID: u.ID,
			Timestamp:   time.Now(),
			Version:     u.Version + 1,
		},
		NewPasswordHash: string(hash),
	}

	u.applyChange(event, true)
	return nil
}

// ChangeEmail - Email değiştirme
func (u *UserAggregate) ChangeEmail(newEmail string) error {
	// Business rule: User aktif olmalı
	if u.Status != "active" {
		return errors.New("user is not active")
	}

	// Business rule: Email boş olmamalı
	if newEmail == "" {
		return errors.New("email cannot be empty")
	}

	// Business rule: Email aynı olmamalı
	if u.Email == newEmail {
		return errors.New("new email must be different from current email")
	}

	// Event oluştur ve uygula
	event := EmailChangedEvent{
		BaseEvent: BaseEvent{
			AggregateID: u.ID,
			Timestamp:   time.Now(),
			Version:     u.Version + 1,
		},
		OldEmail: u.Email,
		NewEmail: newEmail,
	}

	u.applyChange(event, true)
	return nil
}

// Deactivate - User'ı deaktive etme
func (u *UserAggregate) Deactivate(reason string) error {
	// Business rule: User zaten deaktif olmamalı
	if u.Status == "deactivated" {
		return errors.New("user is already deactivated")
	}

	// Event oluştur ve uygula
	event := UserDeactivatedEvent{
		BaseEvent: BaseEvent{
			AggregateID: u.ID,
			Timestamp:   time.Now(),
			Version:     u.Version + 1,
		},
		Reason: reason,
	}

	u.applyChange(event, true)
	return nil
}

// RecordLogin - Login'i kaydet (stateless event)
func (u *UserAggregate) RecordLogin(ipAddress, userAgent string) error {
	// Business rule: User aktif olmalı
	if u.Status != "active" {
		return errors.New("user is not active")
	}

	// Event oluştur ve uygula
	event := UserLoginRecordedEvent{
		BaseEvent: BaseEvent{
			AggregateID: u.ID,
			Timestamp:   time.Now(),
			Version:     u.Version + 1,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	u.applyChange(event, true)
	return nil
}

// VerifyPassword - Password'ü doğrula (query operasyonu)
func (u *UserAggregate) VerifyPassword(password string) error {
	if u.Status != "active" {
		return errors.New("user is not active")
	}

	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// applyChange - Event'i aggregate state'ine uygular
func (u *UserAggregate) applyChange(event DomainEvent, isNew bool) {
	switch e := event.(type) {
	case UserCreatedEvent:
		u.Email = e.Email
		u.PasswordHash = e.PasswordHash
		u.Status = "active"
		u.CreatedAt = e.Timestamp
		u.UpdatedAt = e.Timestamp

	case PasswordChangedEvent:
		u.PasswordHash = e.NewPasswordHash
		u.UpdatedAt = e.Timestamp

	case EmailChangedEvent:
		u.Email = e.NewEmail
		u.UpdatedAt = e.Timestamp

	case UserDeactivatedEvent:
		u.Status = "deactivated"
		u.UpdatedAt = e.Timestamp

	case UserLoginRecordedEvent:
		// Login event aggregate state'ini değiştirmez
		// Sadece event store'a kaydedilir
		u.UpdatedAt = e.Timestamp
	}

	// Version'ı güncelle
	u.Version = event.GetVersion()

	// Eğer yeni bir event ise uncommitted changes'a ekle
	if isNew {
		u.uncommittedChanges = append(u.uncommittedChanges, event)
	}
}
