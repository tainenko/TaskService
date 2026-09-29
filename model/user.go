package model

import "time"

// User is an account that owns tasks.
type User struct {
	ID           int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Email        string    `gorm:"column:email;not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (*User) TableName() string { return "users" }
