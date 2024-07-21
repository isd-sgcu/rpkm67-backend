package test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-backend/config"
	service "github.com/isd-sgcu/rpkm67-backend/internal/selection"
	mock_cache "github.com/isd-sgcu/rpkm67-backend/mocks/cache"
	mock_group "github.com/isd-sgcu/rpkm67-backend/mocks/group"
	mock_selection "github.com/isd-sgcu/rpkm67-backend/mocks/selection"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/selection/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SelectionServiceTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockRepo      *mock_selection.MockRepository
	mockCache     *mock_cache.MockRepository
	mockGroupRepo *mock_group.MockRepository
	service       proto.SelectionServiceServer
	ctx           context.Context
	logger        *zap.Logger
	config        *config.SelectionConfig
}

func TestSelectionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SelectionServiceTestSuite))
}

func (s *SelectionServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mock_selection.NewMockRepository(s.ctrl)
	s.mockCache = mock_cache.NewMockRepository(s.ctrl)
	s.logger = zap.NewNop()
	s.config = &config.SelectionConfig{CacheTTL: 3600}
	s.mockGroupRepo = mock_group.NewMockRepository(s.ctrl)
	s.service = service.NewService(s.mockRepo, s.mockGroupRepo, s.mockCache, s.config, s.logger)
	s.ctx = context.Background()
}

func (s *SelectionServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *SelectionServiceTestSuite) TestCreate_Success() {
	groupID := uuid.New().String()
	baanID := "baan1"
	order := int32(1)

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).Return(nil)
	s.mockRepo.EXPECT().Create(gomock.Any()).Return(nil)

	req := &proto.CreateSelectionRequest{
		GroupId: groupID,
		BaanId:  baanID,
		Order:   order,
	}

	res, err := s.service.Create(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(groupID, res.Selection.GroupId)
	s.Equal(baanID, res.Selection.BaanId)
	s.Equal(order, res.Selection.Order)
}

func (s *SelectionServiceTestSuite) TestCreate_InvalidOrder() {
	groupID := uuid.New().String()
	req := &proto.CreateSelectionRequest{
		GroupId: groupID,
		BaanId:  "baan1",
		Order:   6,
	}

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).Return(nil)

	_, err := s.service.Create(s.ctx, req)

	s.Error(err)
	s.Equal(codes.Internal, status.Code(err))
	s.Contains(err.Error(), "Order must be in range 1-5")
}

func (s *SelectionServiceTestSuite) TestCreate_DuplicateBaan() {
	groupID := uuid.New().String()
	baanID := "baan1"
	parsedUUID := uuid.MustParse(groupID)
	existingSelections := []model.Selection{
		{GroupID: &parsedUUID, Baan: baanID, Order: 1},
	}

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).SetArg(1, existingSelections).Return(nil)

	req := &proto.CreateSelectionRequest{
		GroupId: groupID,
		BaanId:  baanID,
		Order:   2,
	}

	_, err := s.service.Create(s.ctx, req)

	s.Error(err)
	s.Equal(codes.Internal, status.Code(err))
	s.Contains(err.Error(), "Can not create selection with same baan")
}

func (s *SelectionServiceTestSuite) TestCreate_InvalidGroupID() {
	req := &proto.CreateSelectionRequest{
		GroupId: "invalid-uuid",
		BaanId:  "baan1",
		Order:   1,
	}

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	_, err := s.service.Create(s.ctx, req)

	s.Error(err)
	s.Equal(codes.Internal, status.Code(err))
}

func (s *SelectionServiceTestSuite) TestFindByGroupId_Success() {
	groupID := uuid.New().String()
	parsedUUID := uuid.MustParse(groupID)
	selections := []model.Selection{
		{GroupID: &parsedUUID, Baan: "baan1", Order: 1},
		{GroupID: &parsedUUID, Baan: "baan2", Order: 2},
	}

	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).SetArg(1, selections).Return(nil)

	req := &proto.FindByGroupIdSelectionRequest{GroupId: groupID}
	res, err := s.service.FindByGroupId(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Len(res.Selections, 2)
	s.Equal(groupID, res.Selections[0].GroupId)
	s.Equal("baan1", res.Selections[0].BaanId)
	s.Equal(int32(1), res.Selections[0].Order)
}

func (s *SelectionServiceTestSuite) TestDelete_Success() {
	groupID := uuid.New().String()
	baanID := "baan1"

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().Delete(groupID, baanID).Return(nil)

	req := &proto.DeleteSelectionRequest{GroupId: groupID, BaanId: baanID}
	res, err := s.service.Delete(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.True(res.Success)
}

func (s *SelectionServiceTestSuite) TestCountByBaanId_CacheHit() {
	cachedResponse := &proto.CountByBaanIdSelectionResponse{
		BaanCounts: []*proto.BaanCount{
			{BaanId: "baan1", Count: 5},
			{BaanId: "baan2", Count: 3},
		},
	}

	s.mockCache.EXPECT().GetValue("countByBaanId", gomock.Any()).SetArg(1, cachedResponse).Return(nil)

	req := &proto.CountByBaanIdSelectionRequest{}
	res, err := s.service.CountByBaanId(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(cachedResponse.BaanCounts, res.BaanCounts)
}

func (s *SelectionServiceTestSuite) TestCountByBaanId_CacheMiss() {
	count := map[string]int{
		"baan1": 5,
		"baan2": 3,
	}

	s.mockCache.EXPECT().GetValue("countByBaanId", gomock.Any()).Return(status.Error(codes.NotFound, "cache miss"))
	s.mockRepo.EXPECT().CountByBaanId().Return(count, nil)
	s.mockCache.EXPECT().SetValue("countByBaanId", gomock.Any(), s.config.CacheTTL).Return(nil)

	req := &proto.CountByBaanIdSelectionRequest{}
	res, err := s.service.CountByBaanId(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Len(res.BaanCounts, 2)
}

func (s *SelectionServiceTestSuite) TestUpdate_UpdateExistBaanNewOrderSuccess() {
	groupID := uuid.New().String()
	parsedUUID := uuid.MustParse(groupID)
	baanID := "baan1"
	order := int32(2)

	oldSelections := []model.Selection{
		{GroupID: &parsedUUID, Baan: baanID, Order: 1},
	}

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).SetArg(1, oldSelections).Return(nil)
	s.mockRepo.EXPECT().UpdateExistBaanNewOrder(gomock.Any()).Return(nil)

	req := &proto.UpdateSelectionRequest{
		GroupId: groupID,
		BaanId:  baanID,
		Order:   order,
	}

	res, err := s.service.Update(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(groupID, res.Selection.GroupId)
	s.Equal(baanID, res.Selection.BaanId)
	s.Equal(order, res.Selection.Order)
}

func (s *SelectionServiceTestSuite) TestUpdate_UpdateExistBaanExistOrderSuccess() {
	groupID := uuid.New().String()
	parsedUUID := uuid.MustParse(groupID)
	baanID := "baan1"
	order := int32(2)

	oldSelections := []model.Selection{
		{GroupID: &parsedUUID, Baan: baanID, Order: 1},
		{GroupID: &parsedUUID, Baan: "baan2", Order: int(order)},
	}

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).SetArg(1, oldSelections).Return(nil)
	s.mockRepo.EXPECT().UpdateExistBaanExistOrder(gomock.Any()).Return(nil)

	req := &proto.UpdateSelectionRequest{
		GroupId: groupID,
		BaanId:  baanID,
		Order:   order,
	}

	res, err := s.service.Update(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(groupID, res.Selection.GroupId)
	s.Equal(baanID, res.Selection.BaanId)
	s.Equal(order, res.Selection.Order)
}

func (s *SelectionServiceTestSuite) TestUpdate_UpdateNewBaanExistOrderSuccess() {
	groupID := uuid.New().String()
	parsedUUID := uuid.MustParse(groupID)
	baanID := "baan1"
	order := int32(2)

	oldSelections := []model.Selection{
		{GroupID: &parsedUUID, Baan: "baan2", Order: int(order)},
	}

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).SetArg(1, oldSelections).Return(nil)
	s.mockRepo.EXPECT().UpdateNewBaanExistOrder(gomock.Any()).Return(nil)

	req := &proto.UpdateSelectionRequest{
		GroupId: groupID,
		BaanId:  baanID,
		Order:   order,
	}

	res, err := s.service.Update(s.ctx, req)

	s.NoError(err)
	s.NotNil(res)
	s.Equal(groupID, res.Selection.GroupId)
	s.Equal(baanID, res.Selection.BaanId)
	s.Equal(order, res.Selection.Order)
}

func (s *SelectionServiceTestSuite) TestUpdate_InvalidScenario() {
	groupID := uuid.New().String()
	parsedUUID := uuid.MustParse(groupID)
	baanID := "newBaan"
	order := int32(3)

	oldSelections := []model.Selection{
		{GroupID: &parsedUUID, Baan: "baan1", Order: 1},
		{GroupID: &parsedUUID, Baan: "baan2", Order: 2},
	}

	s.mockGroupRepo.EXPECT().FindOne(gomock.Any(), gomock.Any()).SetArg(1, model.Group{IsConfirmed: false}).Return(nil)
	s.mockRepo.EXPECT().FindByGroupId(groupID, gomock.Any()).SetArg(1, oldSelections).Return(nil)

	req := &proto.UpdateSelectionRequest{
		GroupId: groupID,
		BaanId:  baanID,
		Order:   order,
	}

	res, err := s.service.Update(s.ctx, req)

	s.Error(err)
	s.Nil(res)
	s.Equal(codes.Internal, status.Code(err))
	s.Contains(err.Error(), "Invalid update scenario")
}
