package pin

import (
	"context"
	"fmt"

	"github.com/isd-sgcu/rpkm67-backend/internal/dto"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/pin/v1"
	"go.uber.org/zap"
)

type Service interface {
	proto.PinServiceServer
}

type serviceImpl struct {
	proto.UnimplementedPinServiceServer
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) Service {
	return &serviceImpl{
		repo: repo,
		log:  log,
	}
}

func (s *serviceImpl) FindAll(_ context.Context, in *proto.FindAllPinRequest) (res *proto.FindAllPinResponse, err error) {
	res = &proto.FindAllPinResponse{}
	keys := []string{
		"workshop-1",
		"workshop-2",
		"workshop-3",
		"workshop-4",
		"workshop-5",
	}

	for _, key := range keys {
		pin := &dto.Pin{}

		err = s.repo.GetPin(key, pin)
		if err != nil {
			s.log.Named("FindAll").Error(fmt.Sprintf("GetPin: key=%s", key), zap.Error(err))
			if err.Error() != "redis: nil" {
				return nil, err
			}

			code, err := generatePIN()
			if err != nil {
				s.log.Named("FindAll").Error("generatePIN: ", zap.Error(err))
				return nil, err
			}

			err = s.repo.SetPin(key, &dto.Pin{Code: code})
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

	code, err := generatePIN()
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
