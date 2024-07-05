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

	// Try to retrieve from cache
	err := s.cache.GetValue(cacheKey, &cachedGroup)
	if err == nil {
		s.log.Info("Group found in cache", zap.String("user_id", in.UserId))
		return &proto.FindOneGroupResponse{Group: &cachedGroup}, nil
	}

	// If not found in cache, fetch from database
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		s.log.Error("Invalid UUID format", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	// Convert to RPC format and update cache
	groupRPC := s.convertToGroupRPC(group)
	if err := s.cache.SetValue(cacheKey, groupRPC, 3600); err != nil {
		s.log.Warn("Failed to set group in cache", zap.String("user_id", in.UserId), zap.Error(err))
	}

	s.log.Info("FindOne group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("user_id", in.UserId),
		zap.Int("member_count", len(group.Members)),
		zap.Bool("from_cache", false))

	return &proto.FindOneGroupResponse{Group: groupRPC}, nil
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
		return nil, fmt.Errorf("failed to find group by token: %w", err)
	}

	if len(group.Members) == 0 {
		s.log.Error("Group has no members", zap.String("token", in.Token))
		return nil, errors.New("group has no members")
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
	if err := s.cache.SetValue(cacheKey, &res, 3600); err != nil {
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
		s.log.Error("Invalid UUID format", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	_, err = s.repo.FindOne(leaderUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	if err := s.repo.Update(leaderUUID, &model.Group{IsConfirmed: in.Group.IsConfirmed}); err != nil {
		s.log.Error("Failed to update group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	updatedGroup, err := s.repo.FindOne(leaderUUID)
	if err != nil {
		s.log.Error("Failed to find updated group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, fmt.Errorf("failed to find updated group: %w", err)
	}

	// Update cache
	s.updateGroupCache(updatedGroup)

	groupRPC := s.convertToGroupRPC(updatedGroup)

	s.log.Info("UpdateGroup service completed",
		zap.String("group_id", updatedGroup.ID.String()),
		zap.String("leader_id", in.LeaderId),
		zap.Bool("is_confirmed", updatedGroup.IsConfirmed))

	return &proto.UpdateGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) DeleteMember(ctx context.Context, in *proto.DeleteMemberGroupRequest) (*proto.DeleteMemberGroupResponse, error) {
	leaderUUID, err := uuid.Parse(in.LeaderId)
	if err != nil {
		s.log.Error("Invalid leader UUID format", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, fmt.Errorf("invalid leader UUID format: %w", err)
	}

	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		s.log.Error("Invalid user UUID format", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("invalid user UUID format: %w", err)
	}

	group, err := s.repo.FindOne(leaderUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	if in.LeaderId != group.LeaderID {
		s.log.Error("Requested leader_id is not leader of this group", zap.String("leader_id", in.LeaderId))
		return nil, errors.New("requested leader_id is not leader of this group")
	}

	var found bool
	for _, member := range group.Members {
		if member.ID.String() == in.UserId {
			found = true
			break
		}
	}
	if !found {
		s.log.Error("User is not in the group", zap.String("user_id", in.UserId))
		return nil, errors.New("this user_id is not in the group")
	}

	var newGroup *model.Group
	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		createdGroup, err := s.repo.CreateNewGroupWithTX(ctx, tx, userUUID.String())
		if err != nil {
			s.log.Error("Failed to create new group", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.repo.DeleteMemberFromGroupWithTX(ctx, tx, userUUID, createdGroup.ID); err != nil {
			s.log.Error("Failed to delete member from group", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to delete member from group: %w", err)
		}

		// find created group for updating cache of poor user
		newGroup, err = s.repo.FindOne(userUUID)
		if err != nil {
			s.log.Error("Failed to find group", zap.String("leader_id", in.LeaderId), zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		s.log.Error("Transaction failed", zap.Error(err))
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	updatedGroup, err := s.repo.FindOne(leaderUUID)
	if err != nil {
		s.log.Error("Failed to find updated group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, fmt.Errorf("failed to find updated group: %w", err)
	}

	// Update cache
	s.updateGroupCache(updatedGroup)
	s.updateGroupCache(newGroup)

	groupRPC := s.convertToGroupRPC(updatedGroup)

	s.log.Info("DeleteMember service completed",
		zap.String("group_id", updatedGroup.ID.String()),
		zap.String("leader_id", in.LeaderId),
		zap.String("deleted_user_id", in.UserId))

	return &proto.DeleteMemberGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Leave(ctx context.Context, in *proto.LeaveGroupRequest) (*proto.LeaveGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		s.log.Error("Invalid UUID format", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	if in.UserId == group.LeaderID {
		s.log.Error("User is the leader of the group", zap.String("user_id", in.UserId))
		return nil, errors.New("requested user_id is leader of this group so you cannot leave")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		createdGroup, err := s.repo.CreateNewGroupWithTX(ctx, tx, userUUID.String())
		if err != nil {
			s.log.Error("Failed to create new group", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.repo.DeleteMemberFromGroupWithTX(ctx, tx, userUUID, createdGroup.ID); err != nil {
			s.log.Error("Failed to delete member from group", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to delete member from group: %w", err)
		}

		return nil
	})

	if err != nil {
		s.log.Error("Transaction failed", zap.Error(err))
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	leaderExistedGroup := group.LeaderID
	leaderExistedGroupUUID, err := uuid.Parse(leaderExistedGroup)
	if err != nil {
		s.log.Error("Invalid UUID format", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}
	existedGroup, err := s.repo.FindOne(leaderExistedGroupUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	updatedGroup, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find updated group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("failed to find updated group: %w", err)
	}

	// Update cache
	s.updateGroupCache(existedGroup)
	s.updateGroupCache(updatedGroup)

	groupRPC := s.convertToGroupRPC(updatedGroup)

	s.log.Info("Leave service completed",
		zap.String("group_id", updatedGroup.ID.String()),
		zap.String("user_id", in.UserId))

	return &proto.LeaveGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Join(ctx context.Context, in *proto.JoinGroupRequest) (*proto.JoinGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		s.log.Error("Invalid UUID format", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	existedGroup, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find existing group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("failed to find existing group: %w", err)
	}

	if in.UserId == existedGroup.LeaderID && len(existedGroup.Members) > 1 {
		s.log.Error("User is the leader of a non-empty group", zap.String("user_id", in.UserId))
		return nil, errors.New("requested user_id is leader of this group so you cannot leave")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		group, tokenErr := s.repo.FindByToken(in.Token)
		if tokenErr != nil {
			s.log.Error("Failed to find group by token", zap.String("token", in.Token), zap.Error(tokenErr))
			return fmt.Errorf("failed to find group by token: %w", tokenErr)
		}

		err := s.repo.JoinGroupWithTX(ctx, tx, userUUID, group.ID)
		if err != nil {
			s.log.Error("Failed to join group", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to join group: %w", err)
		}

		if in.UserId == existedGroup.LeaderID && len(existedGroup.Members) == 1 {
			err := s.repo.DeleteGroup(ctx, tx, existedGroup.ID)
			if err != nil {
				s.log.Error("Failed to delete old group", zap.String("group_id", existedGroup.ID.String()), zap.Error(err))
				return fmt.Errorf("failed to delete old group: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		s.log.Error("Transaction failed", zap.Error(err))
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	updatedGroup, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find updated group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, fmt.Errorf("failed to find updated group: %w", err)
	}

	// Update cache
	s.updateGroupCache(updatedGroup)

	groupRPC := s.convertToGroupRPC(updatedGroup)

	s.log.Info("Join service completed",
		zap.String("group_id", updatedGroup.ID.String()),
		zap.String("user_id", in.UserId))

	return &proto.JoinGroupResponse{Group: groupRPC}, nil
}

// Helper functions

func (s *serviceImpl) updateGroupCache(group *model.Group) {
	groupRPC := s.convertToGroupRPC(group)
	for _, member := range group.Members {
		cacheKey := fmt.Sprintf("group:%s", member.ID.String())
		if err := s.cache.SetValue(cacheKey, groupRPC, 3600); err != nil {
			s.log.Warn("Failed to update group cache", zap.String("user_id", member.ID.String()), zap.Error(err))
		}
	}
}

func (s *serviceImpl) convertToGroupRPC(group *model.Group) *proto.Group {
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

	return &proto.Group{
		Id:          group.ID.String(),
		LeaderID:    group.LeaderID,
		Token:       group.Token,
		IsConfirmed: group.IsConfirmed,
		Members:     membersRPC,
	}
}
