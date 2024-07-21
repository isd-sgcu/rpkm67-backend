package stamp

import (
	"context"
	"errors"

	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/stamp/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	proto.StampServiceServer
}

type serviceImpl struct {
	proto.UnimplementedStampServiceServer
	repo            Repository
	activityIdToIdx map[string]int
	log             *zap.Logger
}

func NewService(repo Repository, activityIdToIdx map[string]int, log *zap.Logger) Service {
	return &serviceImpl{
		repo:            repo,
		activityIdToIdx: activityIdToIdx,
		log:             log,
	}
}

func (s *serviceImpl) FindByUserId(_ context.Context, in *proto.FindByUserIdStampRequest) (res *proto.FindByUserIdStampResponse, err error) {
	stamp := &model.Stamp{}

	err = s.repo.FindByUserId(in.UserId, stamp)
	if err != nil {
		s.log.Named("FindByUserId").Error("FindByUserId", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.FindByUserIdStampResponse{Stamp: s.modelToProto(stamp)}, nil
}

func (s *serviceImpl) StampByUserId(_ context.Context, in *proto.StampByUserIdRequest) (res *proto.StampByUserIdResponse, err error) {
	stamp := &model.Stamp{}

	err = s.repo.FindByUserId(in.UserId, stamp)
	if err != nil {
		s.log.Named("FindByUserId").Error("FindByUserId", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	actIdx, ok := s.activityIdToIdx[in.ActivityId]
	if !ok {
		return nil, status.Error(codes.Internal, errors.New("invalid Activity ID").Error())
	}

	tempStrStamp := []byte(stamp.Stamp)
	if tempStrStamp[actIdx] == '1' {
		return nil, status.Error(codes.Internal, errors.New("already stamped").Error())
	}

	if actIdx >= 9 {
		ans := &model.Answer{
			ActivityID: in.ActivityId,
			Text:       in.Answer,
		}
		if err := s.repo.CreateAnswer(ans); err != nil {
			s.log.Named("StampByUserId").Error("CreateAnswer", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	tempStrStamp[actIdx] = '1'
	stamp.Stamp = string(tempStrStamp)
	s.addNewScore(stamp, actIdx)

	err = s.repo.StampByUserId(in.UserId, stamp)
	if err != nil {
		s.log.Named("StampByUserId").Error("StampByUserId", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.StampByUserIdResponse{Stamp: s.modelToProto(stamp)}, nil
}

func (s *serviceImpl) modelToProto(stamp *model.Stamp) *proto.Stamp {
	return &proto.Stamp{
		UserId: stamp.UserID.String(),
		PointA: int32(stamp.PointA),
		PointB: int32(stamp.PointB),
		PointC: int32(stamp.PointC),
		PointD: int32(stamp.PointD),
		Stamp:  stamp.Stamp,
	}
}

func (s *serviceImpl) addNewScore(stamp *model.Stamp, idx int) {
	if idx <= 1 {
		stamp.PointB += 2
		stamp.PointD += 2
	} else if idx == 2 || idx == 4 {
		stamp.PointA++
		stamp.PointB++
		stamp.PointC += 2
	} else if idx == 3 {
		stamp.PointB += 2
		stamp.PointC++
		stamp.PointD++
	} else if idx >= 5 && idx <= 7 {
		stamp.PointA++
		stamp.PointD++
	} else if idx == 8 {
		stamp.PointA++
	} else {
		stamp.PointA += 2
	}
}
