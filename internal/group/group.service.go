package group

import (
	"context"

	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/group/v1"
	"go.uber.org/zap"
)

type Service interface {
	proto.GroupServiceServer
}

type serviceImpl struct {
	proto.UnimplementedGroupServiceServer
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) Service {
	return &serviceImpl{
		repo: repo,
		log:  log,
	}
}

func (s *serviceImpl) FindOne(_ context.Context, in *proto.FindOneGroupRequest) (res *proto.FindOneGroupResponse, err error) {
	return nil, nil
}
