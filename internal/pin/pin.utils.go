package pin

import (
	"fmt"

	"golang.org/x/exp/rand"
)

func generatePIN() (string, error) {

	max := 999999
	min := 100000
	pin := rand.Intn(max-min+1) + min

	return fmt.Sprint(pin), nil
}
