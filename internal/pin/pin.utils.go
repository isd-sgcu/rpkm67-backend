package pin

import (
	"fmt"

	"golang.org/x/exp/rand"
)

type Utils interface {
	GeneratePIN() (string, error)
}

type utilsImpl struct{}

func NewUtils() Utils {
	return &utilsImpl{}
}

func (u *utilsImpl) GeneratePIN() (string, error) {

	max := 999999
	min := 100000
	pin := rand.Intn(max-min+1) + min

	return fmt.Sprint(pin), nil
}
