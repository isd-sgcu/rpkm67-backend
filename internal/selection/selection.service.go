package selection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/selection/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
	"go.uber.org/zap"
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
	GroupUuid, err := uuid.Parse(in.GroupId)
	if err != nil {
		s.log.Error("Failed to parse group id", zap.Error(err))
		return nil, err
	}

	selections := &[]model.Selection{}
	err = s.repo.FindByGroupId(in.GroupId, selections)
	if err != nil {
		s.log.Error("Failed to find selection", zap.Error(err))
		return nil, err
	}

	//Check can not create selection with same order
	for _, selection := range *selections {
		if selection.Order == int(in.Order) {
			s.log.Error("Failed to create selection", zap.Error(err))
			return nil, fmt.Errorf("Can not create selection with same order")
		}
	}

	//Check can not create selection with same baan
	for _, selection := range *selections {
		if selection.Baan == in.BaanId {
			s.log.Error("Failed to create selection", zap.Error(err))
			return nil, fmt.Errorf("Can not create selection with same baan")
		}
	}

	//Order must be in range 1-5
	if in.Order < 1 || in.Order > 5 {
		s.log.Error("Failed to create selection", zap.Error(err))
		return nil, fmt.Errorf("Order must be in range 1-5")
	}

	//Create selection
	selection := model.Selection{
		GroupID: &GroupUuid,
		Baan:    in.BaanId,
		Order:   int(in.Order),
	}

	err = s.repo.Create(&selection)
	if err != nil {
		s.log.Error("Failed to create selection", zap.Error(err))
		return nil, err
	}

	defer func() {
		cacheKey := fmt.Sprintf("group:%s", in.GroupId)
		err = s.cache.SetValue(cacheKey, selection, 3600)
		if err != nil {
			s.log.Error("Failed to cache selection", zap.Error(err))
		}
	}()

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
	cacheKey := fmt.Sprintf("group:%s", in.GroupId)
	var cachedSelection []*proto.Selection

	err := s.cache.GetValue(cacheKey, &cachedSelection)
	if err == nil {
		s.log.Info("Group found in cache", zap.String("user_id", in.GroupId))
		return &proto.FindByGroupIdSelectionResponse{Selections: cachedSelection}, nil
	}

	selection := &[]model.Selection{}

	err = s.repo.FindByGroupId(in.GroupId, selection)
	if err != nil {
		s.log.Error("Failed to find selection", zap.String("group_id", in.GroupId), zap.Error(err))
		return nil, err
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
		s.log.Error("Failed to delete selection", zap.Error(err))
		return nil, err
	}

	defer func() {
		cacheKey := fmt.Sprintf("group:%s", in.GroupId)
		err = s.cache.DeleteValue(cacheKey)
		if err != nil {
			s.log.Error("Failed to delete selection from cache", zap.Error(err))
		}
	}()

	s.log.Info("Selection deleted",
		zap.String("group_id", in.GroupId))

	return &proto.DeleteSelectionResponse{Success: true}, nil
}

func (s *serviceImpl) CountByBaanId(ctx context.Context, in *proto.CountByBaanIdSelectionRequest) (*proto.CountByBaanIdSelectionResponse, error) {
	count, err := s.repo.CountByBaanId()
	if err != nil {
		s.log.Error("Failed to count group by baan id", zap.Error(err))
		return nil, err
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

	s.log.Info("Count group by baan id",
		zap.Any("count", count))

	return &res, nil
}

func (s *serviceImpl) Update(ctx context.Context, in *proto.UpdateSelectionRequest) (*proto.UpdateSelectionResponse, error) {
	oldSelections := &[]model.Selection{}

	err := s.repo.FindByGroupId(in.Selection.GroupId, oldSelections)
	if err != nil {
		s.log.Error("Failed to find selection", zap.String("group_id", in.Selection.GroupId), zap.Error(err))
		return nil, err
	}

	groupUUID, err := uuid.Parse(in.Selection.GroupId)
	if err != nil {
		s.log.Error("Failed to parse group id", zap.Error(err))
		return nil, err
	}

	//Order must be in range 1-5
	if in.Selection.Order < 1 || in.Selection.Order > 5 {
		s.log.Error("Failed to create selection", zap.Error(err))
		return nil, fmt.Errorf("Order must be in range 1-5")
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
		updateErr = s.repo.UpdateNewBaanExistOrder(&newSelection, oldSelections)
	} else if baanExists && orderExists {
		updateErr = s.repo.UpdateExistBaanExistOrder(&newSelection, oldSelections)
	} else if baanExists && !orderExists {
		updateErr = s.repo.UpdateExistBaanNewOrder(&newSelection, oldSelections)
	} else {
		s.log.Error("Invalid update scenario", zap.String("group_id", in.Selection.GroupId), zap.String("baan_id", in.Selection.BaanId))
		return nil, err
	}

	if updateErr != nil {
		s.log.Error("Failed to update selection", zap.String("group_id", in.Selection.GroupId), zap.Error(updateErr))
		return nil, updateErr
	}

	res := proto.UpdateSelectionResponse{
		Success: true,
	}

	s.log.Info("Selection updated",
		zap.String("group_id", in.Selection.GroupId),
		zap.String("baan_id", in.Selection.BaanId))

	return &res, nil
}
