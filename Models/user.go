package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserName string `gorm:"not null"`
	Email    string `gorm:"unique"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null;default:user"`
}

type Admin struct {
	gorm.Model
	AdminName string `gorm:"not null"`
	Email     string `gorm:"unique"`
	Password  string `gorm:"not null"`
}

type InvalidErr struct {
	NameError     string
	EmailError    string
	PasswordError string
	RoleError     string
	Err           string
}

type Compare struct {
	UserName string
	Password string
	Role     string
}

type UserDetails struct {
	UserName string
	Email    string
}