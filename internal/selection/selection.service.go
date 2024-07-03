package selection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/selection/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	proto.SelectionServiceServer
}

type serviceImpl struct {
	proto.UnimplementedSelectionServiceServer
	repo  Repository
	cache cache.Repository
	log   *zap.Logger
}

func NewService(repo Repository, cache cache.Repository, log *zap.Logger) Service {
	return &serviceImpl{
		repo:  repo,
		cache: cache,
		log:   log,
	}
}

func (s *serviceImpl) Create(ctx context.Context, in *proto.CreateSelectionRequest) (*proto.CreateSelectionResponse, error) {
	groupUUID, err := uuid.Parse(in.GroupId)
	if err != nil {
		s.log.Named("Create").Error(fmt.Sprintf("Parse group id: %s", in.GroupId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	selections := &[]model.Selection{}
	err = s.repo.FindByGroupId(in.GroupId, selections)
	if err != nil {
		s.log.Named("Create").Error(fmt.Sprintf("FindByGroupId: group_id=%s", in.GroupId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Check can not create selection with same order
	for _, selection := range *selections {
		if selection.Order == int(in.Order) {
			s.log.Named("Create").Error(fmt.Sprintf("Failed to create selection: order=%d", in.Order), zap.Error(err))
			return nil, status.Error(codes.Internal, "Can not create selection with same order")
		}
	}

	//Check can not create selection with same baan
	for _, selection := range *selections {
		if selection.Baan == in.BaanId {
			s.log.Named("Create").Error(fmt.Sprintf("Failed to create selection: baan_id=%s", in.BaanId), zap.Error(err))
			return nil, status.Error(codes.Internal, "Can not create selection with same baan")
		}
	}

	//Order must be in range 1-5
	if in.Order < 1 || in.Order > 5 {
		s.log.Named("Create").Error(fmt.Sprintf("Failed to create selection: order=%d", in.Order), zap.Error(err))
		return nil, status.Error(codes.Internal, "Order must be in range 1-5")
	}

	//Create selection
	selection := model.Selection{
		GroupID: &groupUUID,
		Baan:    in.BaanId,
		Order:   int(in.Order),
	}

	err = s.repo.Create(&selection)
	if err != nil {
		s.log.Named("Create").Error(fmt.Sprintf("Create: group_id=%s, baan_id=%s", in.GroupId, in.BaanId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := proto.CreateSelectionResponse{
		Selection: &proto.Selection{
			Id:      "",
			GroupId: in.GroupId,
			BaanId:  in.BaanId,
			Order:   in.Order,
		},
	}

	s.log.Info("Selection created",
		zap.String("group_id", in.GroupId),
		zap.String("baan_id", in.BaanId))

	return &res, nil
}

func (s *serviceImpl) FindByGroupId(ctx context.Context, in *proto.FindByGroupIdSelectionRequest) (*proto.FindByGroupIdSelectionResponse, error) {
	selection := &[]model.Selection{}

	err := s.repo.FindByGroupId(in.GroupId, selection)
	if err != nil {
		s.log.Named("FindByGroupId").Error(fmt.Sprintf("FindByGroupId: group_id=%s", in.GroupId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	selectionRPC := []*proto.Selection{}
	for _, m := range *selection {
		ss := &proto.Selection{
			Id:      "",
			GroupId: m.GroupID.String(),
			BaanId:  m.Baan,
			Order:   int32(m.Order),
		}
		selectionRPC = append(selectionRPC, ss)
	}

	res := proto.FindByGroupIdSelectionResponse{
		Selections: selectionRPC,
	}

	s.log.Info("Selection found",
		zap.String("group_id", in.GroupId),
		zap.Any("selections", selectionRPC))

	return &res, nil
}

func (s *serviceImpl) Delete(ctx context.Context, in *proto.DeleteSelectionRequest) (*proto.DeleteSelectionResponse, error) {
	err := s.repo.Delete(in.GroupId, in.BaanId)
	if err != nil {
		s.log.Named("Delete").Error(fmt.Sprintf("Delete: group_id=%s, baan_id=%s", in.GroupId, in.BaanId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.log.Info("Selection deleted",
		zap.String("group_id", in.GroupId))

	return &proto.DeleteSelectionResponse{Success: true}, nil
}

func (s *serviceImpl) CountByBaanId(ctx context.Context, in *proto.CountByBaanIdSelectionRequest) (*proto.CountByBaanIdSelectionResponse, error) {
	cachedKey := "countByBaanId"
	var cachedCount *proto.CountByBaanIdSelectionResponse

	err := s.cache.GetValue(cachedKey, &cachedCount)
	if err == nil {
		s.log.Named("CountByBaanId").Info("Count group by baan id found in cache")
		return &proto.CountByBaanIdSelectionResponse{
			BaanCounts: cachedCount.BaanCounts,
		}, nil
	}

	count, err := s.repo.CountByBaanId()
	if err != nil {
		s.log.Named("CountByBaanId").Error("CountByBaanId", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	countRPC := []*proto.BaanCount{}
	for k, v := range count {
		bc := &proto.BaanCount{
			BaanId: k,
			Count:  int32(v),
		}
		countRPC = append(countRPC, bc)
	}

	res := proto.CountByBaanIdSelectionResponse{
		BaanCounts: countRPC,
	}

	if err := s.cache.SetValue(cachedKey, &res, 3600); err != nil {
		s.log.Named("CountByBaanId").Warn("Failed to set count group by baan id in cache", zap.Error(err))
	}

	s.log.Info("Count group by baan id",
		zap.Any("count", count))

	return &res, nil
}

func (s *serviceImpl) Update(ctx context.Context, in *proto.UpdateSelectionRequest) (*proto.UpdateSelectionResponse, error) {
	oldSelections := &[]model.Selection{}

	err := s.repo.FindByGroupId(in.Selection.GroupId, oldSelections)
	if err != nil {
		s.log.Named("Update").Error(fmt.Sprintf("FindByGroupId: group_id=%s", in.Selection.GroupId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	groupUUID, err := uuid.Parse(in.Selection.GroupId)
	if err != nil {
		s.log.Named("Update").Error(fmt.Sprintf("Parse group id: %s", in.Selection.GroupId), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Order must be in range 1-5
	if in.Selection.Order < 1 || in.Selection.Order > 5 {
		s.log.Named("Update").Error(fmt.Sprintf("Failed to update selection: order=%d", in.Selection.Order), zap.Error(err))
		return nil, status.Error(codes.Internal, "Order must be in range 1-5")
	}

	newSelection := model.Selection{
		GroupID: &groupUUID,
		Baan:    in.Selection.BaanId,
		Order:   int(in.Selection.Order),
	}

	// Check if the new Baan exists in oldSelections
	baanExists := false
	orderExists := false
	for _, oldSel := range *oldSelections {
		if oldSel.Baan == newSelection.Baan {
			baanExists = true
		}
		if oldSel.Order == newSelection.Order {
			orderExists = true
		}
	}

	var updateErr error

	if !baanExists && orderExists {
		updateErr = s.repo.UpdateNewBaanExistOrder(&newSelection)
	} else if baanExists && orderExists {
		updateErr = s.repo.UpdateExistBaanExistOrder(&newSelection)
	} else if baanExists && !orderExists {
		updateErr = s.repo.UpdateExistBaanNewOrder(&newSelection)
	} else {
		s.log.Named("Update").Error(fmt.Sprintf("Invalid update scenario: group_id=%s, baan_id=%s", in.Selection.GroupId, in.Selection.BaanId))
		return nil, status.Error(codes.Internal, "Invalid update scenario")
	}

	if updateErr != nil {
		s.log.Named("Update").Error(fmt.Sprintf("Update: group_id=%s, baan_id=%s", in.Selection.GroupId, in.Selection.BaanId), zap.Error(updateErr))
		return nil, status.Error(codes.Internal, updateErr.Error())
	}

	res := proto.UpdateSelectionResponse{
		Success: true,
	}

	s.log.Info("Selection updated",
		zap.String("group_id", in.Selection.GroupId),
		zap.String("baan_id", in.Selection.BaanId))

	return &res, nil
}
