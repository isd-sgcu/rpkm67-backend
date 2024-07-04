package group

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/group/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	proto.GroupServiceServer
}

type serviceImpl struct {
	proto.UnimplementedGroupServiceServer
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

func (s *serviceImpl) FindOne(ctx context.Context, in *proto.FindOneGroupRequest) (*proto.FindOneGroupResponse, error) {
	cacheKey := fmt.Sprintf("group:%s", in.UserId)
	var cachedGroup proto.Group

	// try to retreive from cache
	err := s.cache.GetValue(cacheKey, &cachedGroup)
	if err == nil {
		s.log.Info("Group found in cache", zap.String("user_id", in.UserId))
		return &proto.FindOneGroupResponse{Group: &cachedGroup}, nil
	}

	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	// if not found cache, find group in database
	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, err
	}

	userInfo := make([]*proto.UserInfo, 0, len(group.Members))
	for _, m := range group.Members {
		user := proto.UserInfo{
			Id:        m.ID.String(),
			Firstname: m.Firstname,
			Lastname:  m.Lastname,
			ImageUrl:  m.PhotoUrl,
		}
		userInfo = append(userInfo, &user)
	}

	groupRPC := proto.Group{
		Id:          group.ID.String(),
		LeaderID:    group.LeaderID,
		Token:       group.Token,
		Members:     userInfo,
		IsConfirmed: group.IsConfirmed,
	}

	// set cache
	if err := s.cache.SetValue(cacheKey, &groupRPC, 3600); err != nil { // cache นาน 1 ชั่วโมง
		s.log.Warn("Failed to set group in cache", zap.String("user_id", in.UserId), zap.Error(err))
	}

	res := proto.FindOneGroupResponse{
		Group: &groupRPC,
	}

	s.log.Info("FindOne group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("user_id", in.UserId),
		zap.Int("member_count", len(userInfo)),
		zap.Bool("from_cache", false))

	return &res, nil
}

func (s *serviceImpl) FindByToken(ctx context.Context, in *proto.FindByTokenGroupRequest) (*proto.FindByTokenGroupResponse, error) {
	cacheKey := fmt.Sprintf("group_token:%s", in.Token)
	var cachedGroup proto.FindByTokenGroupResponse

	// Try to retrieve from cache
	err := s.cache.GetValue(cacheKey, &cachedGroup)
	if err == nil {
		s.log.Info("Group found in cache", zap.String("token", in.Token))
		return &cachedGroup, nil
	}

	// If not found in cache, find group in database
	group, err := s.repo.FindByToken(in.Token)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("token", in.Token), zap.Error(err))
		return nil, err
	}

	if len(group.Members) == 0 {
		s.log.Error("Unexpected error", zap.String("token", in.Token), zap.Error(err))
		return nil, err
	}

	leader := group.Members[0]
	leaderInfo := proto.UserInfo{
		Id:        leader.ID.String(),
		Firstname: leader.Firstname,
		Lastname:  leader.Lastname,
		ImageUrl:  leader.PhotoUrl,
	}

	res := proto.FindByTokenGroupResponse{
		Id:     group.ID.String(),
		Token:  group.Token,
		Leader: &leaderInfo,
	}

	// Set cache
	if err := s.cache.SetValue(cacheKey, &res, 3600); err != nil { // Cache for 1 hour
		s.log.Warn("Failed to set group in cache", zap.String("token", in.Token), zap.Error(err))
	}

	s.log.Info("FindByToken group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("token", in.Token),
		zap.Bool("from_cache", false))

	return &res, nil
}

func (s *serviceImpl) Update(ctx context.Context, in *proto.UpdateGroupRequest) (*proto.UpdateGroupResponse, error) {
	leaderUUID, err := uuid.Parse(in.LeaderId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindOne(leaderUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, err
	}

	if err := s.repo.Update(leaderUUID, &model.Group{IsConfirmed: in.Group.IsConfirmed}); err != nil {
		s.log.Error("Failed to update group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, err
	}

	updatedGroup, err := s.repo.FindOne(leaderUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, err
	}

	cacheKey := fmt.Sprintf("group:%s", in.LeaderId)
	if err := s.cache.SetValue(cacheKey, updatedGroup, 3600); err != nil { // cache นาน 1 ชั่วโมง
		s.log.Warn("Failed to set group in cache", zap.String("leader_id", in.LeaderId), zap.Error(err))
	}

	res := proto.UpdateGroupResponse{
		Group: &proto.Group{
			Id:          group.ID.String(),
			LeaderID:    group.LeaderID,
			Token:       group.Token,
			Members:     nil,
			IsConfirmed: in.Group.IsConfirmed,
		},
	}

	s.log.Info("UpdateGroup group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("leader_id", in.LeaderId),
		zap.Bool("is_confirmed", group.IsConfirmed))

	return &res, nil
}

func (s *serviceImpl) DeleteMember(ctx context.Context, in *proto.DeleteMemberGroupRequest) (*proto.DeleteMemberGroupResponse, error) {
	leaderUUID, err := uuid.Parse(in.LeaderId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindOne(leaderUUID)
	if err != nil {
		return nil, err
	}

	if in.LeaderId != group.LeaderID {
		return nil, errors.New("requested leader_id is not leader of this group")
	}

	var found bool = false
	for _, member := range group.Members {
		if member.ID.String() == in.UserId {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("this user_id is not in the group")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		createdGroup, err := s.repo.CreateNewGroupWithTX(ctx, tx, userUUID.String())
		if err != nil {
			return err
		}

		if err := s.repo.DeleteMemberFromGroupWithTX(ctx, tx, userUUID, createdGroup.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	group, err = s.repo.FindOne(leaderUUID)
	if err != nil {
		return nil, err
	}

	membersRPC := make([]*proto.UserInfo, len(group.Members))
	for i, m := range group.Members {
		dto := proto.UserInfo{
			Id:        m.ID.String(),
			Firstname: m.Firstname,
			Lastname:  m.Lastname,
			ImageUrl:  m.PhotoUrl,
		}
		membersRPC[i] = &dto
	}

	groupRPC := proto.Group{
		Id:          group.ID.String(),
		LeaderID:    group.LeaderID,
		Token:       group.Token,
		IsConfirmed: group.IsConfirmed,
		Members:     membersRPC,
	}

	return &proto.DeleteMemberGroupResponse{
		Group: &groupRPC,
	}, nil
}

func (s *serviceImpl) Leave(ctx context.Context, in *proto.LeaveGroupRequest) (*proto.LeaveGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		return nil, err
	}

	if in.UserId == group.LeaderID {
		return nil, errors.New("requested user_id is leader of this group so you cannot leave")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		createdGroup, err := s.repo.CreateNewGroupWithTX(ctx, tx, userUUID.String())
		if err != nil {
			return err
		}

		if err := s.repo.DeleteMemberFromGroupWithTX(ctx, tx, userUUID, createdGroup.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	group, err = s.repo.FindOne(userUUID)
	if err != nil {
		return nil, err
	}

	membersRPC := make([]*proto.UserInfo, len(group.Members))
	for i, m := range group.Members {
		dto := proto.UserInfo{
			Id:        m.ID.String(),
			Firstname: m.Firstname,
			Lastname:  m.Lastname,
			ImageUrl:  m.PhotoUrl,
		}
		membersRPC[i] = &dto
	}

	groupRPC := proto.Group{
		Id:          group.ID.String(),
		LeaderID:    group.LeaderID,
		Token:       group.Token,
		IsConfirmed: group.IsConfirmed,
		Members:     membersRPC,
	}

	return &proto.LeaveGroupResponse{
		Group: &groupRPC,
	}, nil
}

func (s *serviceImpl) Join(ctx context.Context, in *proto.JoinGroupRequest) (*proto.JoinGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	existedGroup, err := s.repo.FindOne(userUUID)
	if err != nil {
		return nil, err
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		group, tokenErr := s.repo.FindByToken(in.Token)
		if tokenErr != nil {
			return tokenErr
		}

		err := s.repo.JoinGroupWithTX(ctx, tx, userUUID, group.ID)
		if err != nil {
			return err
		}

		if err := s.repo.DeleteMemberFromGroupWithTX(ctx, tx, userUUID, existedGroup.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		return nil, err
	}

	membersRPC := make([]*proto.UserInfo, len(group.Members))
	for i, m := range group.Members {
		dto := proto.UserInfo{
			Id:        m.ID.String(),
			Firstname: m.Firstname,
			Lastname:  m.Lastname,
			ImageUrl:  m.PhotoUrl,
		}
		membersRPC[i] = &dto
	}

	groupRPC := proto.Group{
		Id:          group.ID.String(),
		LeaderID:    group.LeaderID,
		Token:       group.Token,
		IsConfirmed: group.IsConfirmed,
		Members:     membersRPC,
	}

	return &proto.JoinGroupResponse{
		Group: &groupRPC,
	}, nil
}
