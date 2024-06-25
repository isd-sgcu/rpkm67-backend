package token

import "github.com/google/uuid"

type TokenUtils interface {
	GetNewUUID() *uuid.UUID
}

type tokenUtilsImpl struct{}

func NewTokenUtils() TokenUtils {
	return &tokenUtilsImpl{}
}

func (u *tokenUtilsImpl) GetNewUUID() *uuid.UUID {
	uuid := uuid.New()
	return &uuid
}
