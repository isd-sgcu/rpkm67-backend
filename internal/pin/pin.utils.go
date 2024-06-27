package pin

import (
	"fmt"

	"github.com/isd-sgcu/rpkm67-backend/config"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/pin/v1"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

type Utils interface {
	GeneratePIN() (string, error)
}

type utilsImpl struct {
	proto.UnimplementedPinServiceServer
	conf *config.PinConfig
	repo Repository
	log  *zap.Logger
}

func NewUtils() Utils {
	return &utilsImpl{}
}

func (u *utilsImpl) GeneratePIN() (string, error) {

	max := 999999
	min := 100000
	pin := rand.Intn(max-min+1) + min

	return fmt.Sprint(pin), nil
}
