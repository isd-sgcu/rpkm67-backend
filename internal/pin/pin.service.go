package pin

import (
	"context"
	"fmt"

	"github.com/isd-sgcu/rpkm67-backend/config"
	"github.com/isd-sgcu/rpkm67-backend/internal/dto"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/pin/v1"
	"go.uber.org/zap"
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
		pin := &dto.Pin{}

		err = s.repo.GetPin(key, pin)
		if err != nil {
			s.log.Named("FindAll").Error(fmt.Sprintf("GetPin: key=%s", key), zap.Error(err))
			if err.Error() != "redis: nil" {
				return nil, err
			}

			code, err := s.utils.GeneratePIN()
			if err != nil {
				s.log.Named("FindAll").Error("generatePIN: ", zap.Error(err))
				return nil, err
			}

			pin.Code = code

			err = s.repo.SetPin(key, &dto.Pin{Code: pin.Code})
			if err != nil {
				s.log.Named("FindAll").Error(fmt.Sprintf("SetPin: key=%s", key), zap.Error(err))
				return nil, err
			}
		}

		res.Pins = append(res.Pins, &proto.Pin{
			WorkshopId: key,
			Code:       pin.Code,
		})
	}

	return &proto.FindAllPinResponse{
		Pins: res.Pins,
	}, nil
}

func (s *serviceImpl) ResetPin(_ context.Context, in *proto.ResetPinRequest) (res *proto.ResetPinResponse, err error) {
	err = s.repo.DeletePin(in.WorkshopId)
	if err != nil {
		s.log.Named("ResetPin").Error("DeletePin: ", zap.Error(err))
		return nil, err
	}

	code, err := s.utils.GeneratePIN()
	if err != nil {
		s.log.Named("ResetPin").Error("generatePIN: ", zap.Error(err))
		return nil, err
	}
	err = s.repo.SetPin(in.WorkshopId, &dto.Pin{Code: code})
	if err != nil {
		s.log.Named("ResetPin").Error("SetPin: ", zap.Error(err))
		return nil, err
	}

	return &proto.ResetPinResponse{
		Success: true,
	}, nil
}
