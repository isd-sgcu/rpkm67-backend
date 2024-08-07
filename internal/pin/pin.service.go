package pin

import (
	"context"
	"fmt"

	"github.com/isd-sgcu/rpkm67-backend/config"
	"github.com/isd-sgcu/rpkm67-backend/internal/dto"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/pin/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	proto.PinServiceServer
}

type serviceImpl struct {
	proto.UnimplementedPinServiceServer
	conf  *config.PinConfig
	repo  Repository
	utils Utils
	log   *zap.Logger
}

func NewService(conf *config.PinConfig, utils Utils, repo Repository, log *zap.Logger) Service {
	return &serviceImpl{
		conf:  conf,
		repo:  repo,
		utils: utils,
		log:   log,
	}
}

func (s *serviceImpl) FindAll(_ context.Context, in *proto.FindAllPinRequest) (res *proto.FindAllPinResponse, err error) {
	res = &proto.FindAllPinResponse{}
	keys := []string{}

	for i := 1; i <= s.conf.WorkshopCount; i++ {
		keys = append(keys, fmt.Sprintf("%s-%d", s.conf.WorkshopCode, i))
	}
	for i := 1; i <= s.conf.LandmarkCount; i++ {
		keys = append(keys, fmt.Sprintf("%s-%d", s.conf.LandmarkCode, i))
	}

	for _, key := range keys {
		pin, err := s.getPin(key)
		if err != nil {
			s.log.Named("FindAll").Error(fmt.Sprintf("getPin: key=%s", key), zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		res.Pins = append(res.Pins, &proto.Pin{
			ActivityId: key,
			Code:       pin.Code,
		})
	}

	return &proto.FindAllPinResponse{
		Pins: res.Pins,
	}, nil
}

func (s *serviceImpl) ResetPin(_ context.Context, in *proto.ResetPinRequest) (res *proto.ResetPinResponse, err error) {
	err = s.repo.DeletePin(in.ActivityId)
	if err != nil {
		s.log.Named("ResetPin").Error("DeletePin: ", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	code, err := s.utils.GeneratePIN()
	if err != nil {
		s.log.Named("ResetPin").Error("generatePIN: ", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = s.repo.SetPin(in.ActivityId, &dto.Pin{Code: code})
	if err != nil {
		s.log.Named("ResetPin").Error("SetPin: ", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.ResetPinResponse{
		Pin: &proto.Pin{
			ActivityId: in.ActivityId,
			Code:       code,
		},
	}, nil
}

func (s *serviceImpl) CheckPin(_ context.Context, in *proto.CheckPinRequest) (*proto.CheckPinResponse, error) {
	pin, err := s.getPin(in.ActivityId)
	if err != nil {
		s.log.Named("CheckPin").Error(fmt.Sprintf("getPin: key=%s", in.ActivityId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if pin.Code != in.Code {
		return &proto.CheckPinResponse{
			IsMatch: false,
		}, nil
	}

	return &proto.CheckPinResponse{
		IsMatch: true,
	}, nil
}

func (s *serviceImpl) getPin(key string) (*dto.Pin, error) {
	pin := &dto.Pin{}

	err := s.repo.GetPin(key, pin)
	if err != nil {
		s.log.Named("GetPin").Error(fmt.Sprintf("GetPin: key=%s", key), zap.Error(err))
		if err.Error() != "redis: nil" {
			return nil, err
		}

		code, err := s.utils.GeneratePIN()
		if err != nil {
			s.log.Named("GetPin").Error("generatePIN: ", zap.Error(err))
			return nil, err
		}

		pin.Code = code

		err = s.repo.SetPin(key, &dto.Pin{Code: pin.Code})
		if err != nil {
			s.log.Named("GetPin").Error(fmt.Sprintf("SetPin: key=%s", key), zap.Error(err))
			return nil, err
		}
	}

	return pin, nil
}
