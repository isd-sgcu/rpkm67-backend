package group

import (
	"time"

	"github.com/isd-sgcu/rpkm67-backend/internal/utils"
)

func GenToken(id string) string {
	return utils.Hash([]byte(time.Now().String() + id))
}
