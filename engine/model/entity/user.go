package entity

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username       string
	HashedPassword string
	Role           string
}

func NewUser(username string, password string, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := &User{
		Username:       username,
		HashedPassword: string(hashedPassword),
		Role:           role,
	}

	return user, nil
}

func (s *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(s.HashedPassword), []byte(password))
	return err == nil
}

func (s *User) Clone() *User {
	return &User{
		Username:       s.Username,
		HashedPassword: s.HashedPassword,
		Role:           s.Role,
	}
}
