package pin

import (
	"context"

	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/pin/v1"
	"go.uber.org/zap"
)

type Service interface {
	proto.PinServiceServer
}

type serviceImpl struct {
	proto.UnimplementedPinServiceServer
	cache *cache.Repository
	log   *zap.Logger
}

func NewService(cache *cache.Repository, log *zap.Logger) Service {
	return &serviceImpl{
		cache: cache,
		log:   log,
	}
}

func (s *serviceImpl) FindAll(_ context.Context, in *proto.FindAllPinRequest) (res *proto.FindAllPinResponse, err error) {
	return nil, nil
}

func (s *serviceImpl) ResetPin(_ context.Context, in *proto.ResetPinRequest) (res *proto.ResetPinResponse, err error) {
	return nil, nil
}
