package models

import (
	"errors"
	"movies-backend/utils/token"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint   `gorm:"primary_key" json:"id"`
	Email     string `gorm:"size:255;not null;unique" json:"email"`
	Password  string `gorm:"size:255;not null;" json:"password"`
	FirstName string `gorm:"size:255;not null;" json:"first_name"`
	LastName  string `gorm:"size:255;not null;" json:"last_name"`
}

func GetUserByID(uid uint) (User, error) {

	var u User

	if err := DB.First(&u, uid).Error; err != nil {
		return u, errors.New("user not found")
	}

	u.PrepareGive()

	return u, nil

}

func (u *User) PrepareGive() {
	u.Password = ""
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func LoginCheck(email string, password string) (string, User, error) {
	u := User{}

	if err := DB.Model(User{}).Where("email = ?", email).Take(&u).Error; err != nil {
		return "", u, err
	}

	if err := VerifyPassword(password, u.Password); err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", u, err
	}

	jwt, err := token.GenerateToken(u.ID)

	if err != nil {
		return "", u, err
	}

	return jwt, u, nil
}
