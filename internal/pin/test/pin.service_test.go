package test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/isd-sgcu/rpkm67-backend/config"
	"github.com/isd-sgcu/rpkm67-backend/internal/dto"
	"github.com/isd-sgcu/rpkm67-backend/internal/pin"
	mock_pin "github.com/isd-sgcu/rpkm67-backend/mocks/pin"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/pin/v1"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type PinServiceTest struct {
	suite.Suite
	controller *gomock.Controller
	logger     *zap.Logger
	conf       *config.PinConfig
}

func TestPinService(t *testing.T) {
	suite.Run(t, new(PinServiceTest))
}

func (t *PinServiceTest) SetupTest() {
	t.controller = gomock.NewController(t.T())
	t.logger = zap.NewNop()
	t.conf = &config.PinConfig{
		WorkshopCount: 1,
		WorkshopCode:  "workshop",
		LandmarkCount: 0,
		LandmarkCode:  "landmark",
	}
}

func (t *PinServiceTest) TestFindAllSuccess() {
	repo := mock_pin.NewMockRepository(t.controller)
	utils := mock_pin.NewMockUtils(t.controller)
	svc := pin.NewService(t.conf, utils, repo, t.logger)

	expectedResp := &proto.FindAllPinResponse{
		Pins: []*proto.Pin{
			{Code: "123456", ActivityId: "workshop-1"},
		},
	}

	repo.EXPECT().GetPin("workshop-1", &dto.Pin{}).SetArg(1, dto.Pin{Code: "123456"})

	res, err := svc.FindAll(context.Background(), &proto.FindAllPinRequest{})
	t.Equal(res, expectedResp)
	t.Nil(err)
}

func (t *PinServiceTest) TestFindAllNotEmptyError() {
	repo := mock_pin.NewMockRepository(t.controller)
	utils := mock_pin.NewMockUtils(t.controller)
	svc := pin.NewService(t.conf, utils, repo, t.logger)

	repo.EXPECT().GetPin("workshop-1", &dto.Pin{}).Return(errors.New("some error that is not empty error"))

	res, err := svc.FindAll(context.Background(), &proto.FindAllPinRequest{})
	t.Nil(res)
	t.NotNil(err)
}

func (t *PinServiceTest) TestFindAllEmptyError() {
	repo := mock_pin.NewMockRepository(t.controller)
	utils := mock_pin.NewMockUtils(t.controller)
	svc := pin.NewService(t.conf, utils, repo, t.logger)

	expectedResp := &proto.FindAllPinResponse{
		Pins: []*proto.Pin{
			{Code: "111111", ActivityId: "workshop-1"},
		},
	}

	repo.EXPECT().GetPin("workshop-1", &dto.Pin{}).Return(errors.New("redis: nil"))
	utils.EXPECT().GeneratePIN().Return("111111", nil)
	repo.EXPECT().SetPin("workshop-1", &dto.Pin{Code: "111111"}).Return(nil)

	res, err := svc.FindAll(context.Background(), &proto.FindAllPinRequest{})
	t.Equal(expectedResp, res)
	t.Nil(err)
}

func (t *PinServiceTest) TestFindMultiplePins() {
	conf := &config.PinConfig{
		WorkshopCount: 1,
		WorkshopCode:  "workshop",
		LandmarkCount: 1,
		LandmarkCode:  "landmark",
	}
	repo := mock_pin.NewMockRepository(t.controller)
	utils := mock_pin.NewMockUtils(t.controller)
	svc := pin.NewService(conf, utils, repo, t.logger)

	expectedResp := &proto.FindAllPinResponse{
		Pins: []*proto.Pin{
			{Code: "123456", ActivityId: "workshop-1"},
			{Code: "654321", ActivityId: "landmark-1"},
		},
	}

	repo.EXPECT().GetPin("workshop-1", &dto.Pin{}).SetArg(1, dto.Pin{Code: "123456"})
	repo.EXPECT().GetPin("landmark-1", &dto.Pin{}).SetArg(1, dto.Pin{Code: "654321"})

	res, err := svc.FindAll(context.Background(), &proto.FindAllPinRequest{})
	t.Equal(res, expectedResp)
	t.Nil(err)
}

func (t *PinServiceTest) TestResetPinSuccess() {
}
