package command

import "time"

// Command interface - Tüm command'lar bunu implement eder
type Command interface {
	GetAggregateID() string
	GetCommandType() string
}

// RegisterUserCommand - Yeni user kaydı için command
type RegisterUserCommand struct {
	UserID   string
	Email    string
	Password string
}

func (c RegisterUserCommand) GetAggregateID() string {
	return c.UserID
}

func (c RegisterUserCommand) GetCommandType() string {
	return "RegisterUser"
}

// ChangePasswordCommand - Şifre değiştirme command
type ChangePasswordCommand struct {
	UserID      string
	OldPassword string
	NewPassword string
}

func (c ChangePasswordCommand) GetAggregateID() string {
	return c.UserID
}

func (c ChangePasswordCommand) GetCommandType() string {
	return "ChangePassword"
}

// ChangeEmailCommand - Email değiştirme command
type ChangeEmailCommand struct {
	UserID   string
	NewEmail string
}

func (c ChangeEmailCommand) GetAggregateID() string {
	return c.UserID
}

func (c ChangeEmailCommand) GetCommandType() string {
	return "ChangeEmail"
}

// DeactivateUserCommand - User'ı deaktive etme command
type DeactivateUserCommand struct {
	UserID    string
	Reason    string
	Timestamp time.Time
}

func (c DeactivateUserCommand) GetAggregateID() string {
	return c.UserID
}

func (c DeactivateUserCommand) GetCommandType() string {
	return "DeactivateUser"
}

// RecordLoginCommand - Login kaydı için command
type RecordLoginCommand struct {
	UserID    string
	IPAddress string
	UserAgent string
	Timestamp time.Time
}

func (c RecordLoginCommand) GetAggregateID() string {
	return c.UserID
}

func (c RecordLoginCommand) GetCommandType() string {
	return "RecordLogin"
}
