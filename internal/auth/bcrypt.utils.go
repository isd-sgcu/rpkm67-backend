package auth

import "golang.org/x/crypto/bcrypt"

type BcryptUtils interface {
	GenerateHashedPassword(password string) (string, error)
	CompareHashedPassword(hashedPassword string, plainPassword string) error
}

type bcryptUtilsImpl struct{}

func NewBcryptUtils() BcryptUtils {
	return &bcryptUtilsImpl{}
}

func (u *bcryptUtilsImpl) GenerateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func (u *bcryptUtilsImpl) CompareHashedPassword(hashedPassword string, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
