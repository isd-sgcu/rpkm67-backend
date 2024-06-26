package group

import (
	"context"

	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/group/v1"
	"go.uber.org/zap"
)

type Service interface {
	proto.GroupServiceServer
}

type serviceImpl struct {
	proto.UnimplementedGroupServiceServer
	cache *cache.Repository
	log   *zap.Logger
}

func NewService(cache *cache.Repository, log *zap.Logger) Service {
	return &serviceImpl{
		cache: cache,
		log:   log,
	}
}

func (s *serviceImpl) FindOne(_ context.Context, in *proto.FindOneGroupRequest) (res *proto.FindOneGroupResponse, err error) {
	return nil, nil
}
