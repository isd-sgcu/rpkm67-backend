package model

import constants "github.com/isd-sgcu/rpkm67-auth/constant"

type User struct {
	Base
	Email     string         `json:"email" gorm:"tinytext;unique"`
	Password  string         `json:"password" gorm:"tinytext"`
	Firstname string         `json:"firstname" gorm:"tinytext"`
	Lastname  string         `json:"lastname" gorm:"tinytext"`
	Tel       string         `json:"tel" gorm:"tinytext"`
	Role      constants.Role `json:"role" gorm:"tinytext"`
}
