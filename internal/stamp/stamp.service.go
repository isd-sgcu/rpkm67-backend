package stamp

import (
	"context"

	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/stamp/v1"
	"go.uber.org/zap"
)

type Service interface {
	proto.StampServiceServer
}

type serviceImpl struct {
	proto.UnimplementedStampServiceServer
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) Service {
	return &serviceImpl{
		repo: repo,
		log:  log,
	}
}

func (s *serviceImpl) FindByUserId(_ context.Context, in *proto.FindByUserIdStampRequest) (res *proto.FindByUserIdStampResponse, err error) {
	return nil, nil
}

func (s *serviceImpl) StampByUserId(_ context.Context, in *proto.StampByUserIdRequest) (res *proto.StampByUserIdResponse, err error) {
	return nil, nil
}
