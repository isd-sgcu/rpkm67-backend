package dto

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/isd-sgcu/rpkm67-backend/constant"
)

type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthPayload struct {
	jwt.RegisteredClaims
	UserId string        `json:"user_id"`
	Role   constant.Role `json:"role"`
}

type RefreshTokenCache struct {
	UserID string        `json:"user_id"`
	Role   constant.Role `json:"role"`
}

type ResetPasswordTokenCache struct {
	UserID string `json:"user_id"`
}
