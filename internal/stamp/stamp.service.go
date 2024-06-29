package stamp

import (
	"context"

	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/stamp/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
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
	stamp := &model.Stamp{}

	err = s.repo.FindByUserId(in.UserId, stamp)
	if err != nil {
		s.log.Named("FindByUserId").Error("FindByUserId", zap.Error(err))
		return nil, err
	}

	return &proto.FindByUserIdStampResponse{Stamp: s.modelToProto(stamp)}, nil
}

func (s *serviceImpl) StampByUserId(_ context.Context, in *proto.StampByUserIdRequest) (res *proto.StampByUserIdResponse, err error) {
	stamp := &model.Stamp{}

	err = s.repo.FindByUserId(in.UserId, stamp)
	if err != nil {
		s.log.Named("FindByUserId").Error("FindByUserId", zap.Error(err))
		return nil, err
	}

	// stamp.Stamp string 00000000000, 01000100001
	// in.ActivityId = "landmark-4"

	err = s.repo.StampByUserId(in.UserId, stamp)
	if err != nil {
		s.log.Named("StampByUserId").Error("StampByUserId", zap.Error(err))
		return nil, err
	}

	return &proto.StampByUserIdResponse{Stamp: s.modelToProto(stamp)}, nil
}

func (s *serviceImpl) modelToProto(stamp *model.Stamp) *proto.Stamp {
	return &proto.Stamp{
		UserId: stamp.User.ID.String(),
		PointA: int32(stamp.PointA),
		PointB: int32(stamp.PointB),
		PointC: int32(stamp.PointC),
		PointD: int32(stamp.PointD),
		Stamp:  stamp.Stamp,
	}
}
