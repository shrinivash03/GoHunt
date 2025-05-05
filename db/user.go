package db

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string     `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Email     string     `gorm:"unique" json:"email"`
	Password  string     `json:"-"`
	IsAdmin   bool       `gorm:"default:false" json:"isAdmin"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func CreateAdmin(email string, password string) (User, error) {
	admin := User{
		Email: email,
		Password: password,
		IsAdmin: true,
	}
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), 14)
	if err != nil {
		return User{}, errors.New("error creating password")
	}
	admin.Password = string(hashedPassword)

	// Create User in DB
	if err := DBconn.Create(&admin).Error; err != nil {
		return User{}, errors.New("error creating user")
	}
	return admin, nil
}

func (user *User) LoginAsAdmin(email string, password string) (*User, error) {
	// find the user
	if err := DBconn.Where("email = ? AND is_admin = ?", email, true).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}
	
	// compare the passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return user, nil
}