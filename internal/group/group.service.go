package group

import (
	"context"
	"fmt"

	"github.com/isd-sgcu/rpkm67-backend/config"
	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/group/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Service interface {
	proto.GroupServiceServer
}

type serviceImpl struct {
	proto.UnimplementedGroupServiceServer
	repo  Repository
	cache cache.Repository
	conf  *config.GroupConfig
	log   *zap.Logger
}

func NewService(repo Repository, cache cache.Repository, conf *config.GroupConfig, log *zap.Logger) Service {
	return &serviceImpl{
		repo:  repo,
		cache: cache,
		log:   log,
	}
}

// func (s *serviceImpl) FindOne(_ context.Context, in *proto.FindOneGroupRequest) (*proto.FindOneGroupResponse, error) {
// 	cacheKey := groupKey(in.Id)
// 	var cachedGroup proto.Group

// 	if err := s.cache.GetValue(cacheKey, &cachedGroup); err == nil {
// 		s.log.Named("FindOne").Info("GetValue: Group found in cache", zap.String("id", in.Id))
// 		return &proto.FindOneGroupResponse{Group: &cachedGroup}, nil
// 	}

// 	// If not found in cache, fetch from database
// 	group := &model.Group{}
// 	if err := s.repo.FindOne(in.Id, group); err != nil {
// 		s.log.Named("FindOne").Error("FindOne: ", zap.Error(err))
// 		return nil, status.Error(codes.Internal, "failed to find group")
// 	}

// 	groupRPC := ModelToProto(group)
// 	if err := s.cache.SetValue(cacheKey, groupRPC, 3600); err != nil {
// 		s.log.Named("FindOne").Error("SetValue: ", zap.Error(err))
// 		return nil, status.Error(codes.Internal, "failed to cache group")
// 	}

// 	return &proto.FindOneGroupResponse{Group: groupRPC}, nil
// }

func (s *serviceImpl) FindByToken(_ context.Context, in *proto.FindByTokenGroupRequest) (*proto.FindByTokenGroupResponse, error) {
	cacheKey := groupByTokenKey(in.Token)
	var cachedGroup proto.FindByTokenGroupResponse
	if err := s.cache.GetValue(cacheKey, &cachedGroup); err == nil {
		s.log.Named("FindByToken").Info("GetValue: Group found in cache", zap.String("token", in.Token))
		return &cachedGroup, nil
	}

	// If not found in cache, find group in database
	group := &model.Group{}
	if err := s.repo.FindByToken(in.Token, group); err != nil {
		s.log.Named("FindByToken").Error("FindByToken: ", zap.String("token", in.Token), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group by token")
	}

	if err := s.checkGroup(group); err != nil {
		s.log.Named("FindByToken").Error("checkGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "group failed validation")
	}

	var leader *model.User
	for _, member := range group.Members {
		if member.ID == *group.LeaderID {
			leader = member
			break
		}
	}
	leaderInfo := UserToUserInfo(leader)

	res := proto.FindByTokenGroupResponse{
		Id:     group.ID.String(),
		Token:  group.Token,
		Leader: leaderInfo,
	}

	if err := s.cache.SetValue(cacheKey, &res, s.conf.CacheTTL); err != nil {
		s.log.Named("FindByToken").Error("SetValue: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to cache group")
	}

	return &res, nil
}

func (s *serviceImpl) UpdateConfirm(_ context.Context, in *proto.UpdateConfirmGroupRequest) (*proto.UpdateConfirmGroupResponse, error) {
	group := &model.Group{}
	if err := s.repo.FindByUserId(in.LeaderId, group); err != nil {
		s.log.Named("Update").Error("FindOne group: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	if err := s.checkGroup(group); err != nil {
		s.log.Named("Update").Error("checkGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "group failed validation")
	}

	if group.LeaderID.String() != in.LeaderId {
		s.log.Named("Update").Error("Requested leader_id is not leader of this group", zap.String("leader_id", in.LeaderId))
		return nil, status.Error(codes.PermissionDenied, "requested leader_id is not leader of this group")
	}

	group.IsConfirmed = in.IsConfirmed
	if err := s.repo.Update(group.ID.String(), group); err != nil {
		s.log.Named("Update").Error("Update: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update group")
	}

	if err := s.updateGroupCacheByUserId(group); err != nil {
		s.log.Named("Update").Error("updateGroupCacheByUserId: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update group cache")
	}
	groupRPC := ModelToProto(group)

	return &proto.UpdateConfirmGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) DeleteMember(_ context.Context, in *proto.DeleteMemberGroupRequest) (*proto.DeleteMemberGroupResponse, error) {
	group := &model.Group{}
	if err := s.repo.FindByUserId(in.UserId, group); err != nil {
		s.log.Named("DeleteMember").Error("FindOne: ", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	if in.LeaderId != group.LeaderID.String() {
		s.log.Named("DeleteMember").Error("Requested leader_id is not leader of this group", zap.String("leader_id", in.LeaderId))
		return nil, status.Error(codes.PermissionDenied, "requested leader_id is not leader of this group")
	}

	var found bool
	for _, member := range group.Members {
		if member.ID.String() == in.UserId {
			found = true
			break
		}
	}
	if !found {
		s.log.Named("DeleteMember").Error("User is not in the group", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.NotFound, "requested user_id is not in the group")
	}

	err := s.repo.WithTransaction(func(tx *gorm.DB) error {
		createGroup := &model.Group{
			LeaderID: group.LeaderID,
		}

		if err := s.repo.CreateTX(tx, createGroup); err != nil {
			s.log.Named("DeleteMember").Error("CreateTX: ", zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.repo.MoveUserToNewGroupTX(tx, in.UserId, &createGroup.ID); err != nil {
			s.log.Named("DeleteMember").Error("MoveUserToNewGroupTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to delete member from group: %w", err)
		}

		return nil
	})

	if err != nil {
		s.log.Named("DeleteMember").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	newGroup := &model.Group{}
	if err := s.repo.FindByUserId(in.UserId, newGroup); err != nil {
		s.log.Named("DeleteMember").Error("FindOne newGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find new group")
	}
	updatedGroup := &model.Group{}
	if err := s.repo.FindByUserId(in.LeaderId, updatedGroup); err != nil {
		s.log.Named("DeleteMember").Error("FindOne updatedGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find updated group")
	}

	if err := s.updateGroupCacheByUserId(newGroup); err != nil {
		s.log.Named("DeleteMember").Error("updateGroupCacheByUserId: newGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update newGroup cache")
	}
	if err := s.updateGroupCacheByUserId(updatedGroup); err != nil {
		s.log.Named("DeleteMember").Error("updateGroupCacheByUserId: updatedGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update updatedGroup cache")
	}

	groupRPC := ModelToProto(updatedGroup)

	return &proto.DeleteMemberGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Leave(_ context.Context, in *proto.LeaveGroupRequest) (*proto.LeaveGroupResponse, error) {
	group := &model.Group{}
	if err := s.repo.FindByUserId(in.UserId, group); err != nil {
		s.log.Named("Leave").Error("FindOne group: ", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	if in.UserId == group.LeaderID.String() {
		s.log.Named("Leave").Error("User is the leader of the group", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "You are the group leader, so you cannot leave")
	}

	err := s.repo.WithTransaction(func(tx *gorm.DB) error {
		createGroup := &model.Group{
			LeaderID: group.LeaderID,
		}

		if err := s.repo.CreateTX(tx, createGroup); err != nil {
			s.log.Named("Leave").Error("CreateTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.repo.MoveUserToNewGroupTX(tx, in.UserId, &createGroup.ID); err != nil {
			s.log.Named("Leave").Error("MoveUserToNewGroupTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to delete member from group: %w", err)
		}

		return nil
	})

	if err != nil {
		s.log.Named("Leave").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	newGroup := &model.Group{}
	if err := s.repo.FindByUserId(in.UserId, newGroup); err != nil {
		s.log.Named("DeleteMember").Error("FindOne newGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find new group")
	}
	updatedGroup := &model.Group{}
	if err := s.repo.FindByUserId(group.LeaderID.String(), updatedGroup); err != nil {
		s.log.Named("DeleteMember").Error("FindOne updatedGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find updated group")
	}

	if err := s.updateGroupCacheByUserId(newGroup); err != nil {
		s.log.Named("Leave").Error("updateGroupCacheByUserId: newGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update newGroup cache")
	}
	if err := s.updateGroupCacheByUserId(updatedGroup); err != nil {
		s.log.Named("Leave").Error("updateGroupCacheByUserId: updatedGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update updatedGroup cache")
	}

	groupRPC := ModelToProto(updatedGroup)

	return &proto.LeaveGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Join(_ context.Context, in *proto.JoinGroupRequest) (*proto.JoinGroupResponse, error) {
	prevGroup := &model.Group{}
	if err := s.repo.FindByUserId(in.UserId, prevGroup); err != nil {
		s.log.Named("Join").Error("FindByUserId prevGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find previous group of user")
	}

	if in.UserId == prevGroup.LeaderID.String() && len(prevGroup.Members) > 1 {
		s.log.Named("Join").Error("User is the leader of a group with >1 members", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "You are the group leader, so you must kick all other members before joining another group")
	}

	err := s.repo.WithTransaction(func(tx *gorm.DB) error {
		joiningGroup := &model.Group{}
		if err := s.repo.FindByToken(in.Token, joiningGroup); err != nil {
			s.log.Named("Join").Error("FindByToken joiningGroup TX: ", zap.Error(err))
			return fmt.Errorf("failed to find group by token: %w", err)
		}

		if err := s.repo.JoinGroupTX(tx, in.UserId, &joiningGroup.ID); err != nil {
			s.log.Named("Join").Error("JoinGroupTX: ", zap.Error(err))
			return fmt.Errorf("failed to join group: %w", err)
		}

		if in.UserId == prevGroup.LeaderID.String() && len(prevGroup.Members) == 1 {
			err := s.repo.DeleteGroupTX(tx, &prevGroup.ID)
			if err != nil {
				s.log.Named("Join").Error("DeleteGroupTX: ", zap.Error(err))
				return fmt.Errorf("failed to delete old group: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		s.log.Named("Join").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	joiningGroup := &model.Group{}
	if err := s.repo.FindByToken(in.Token, joiningGroup); err != nil {
		s.log.Named("Join").Error("FindByToken joiningGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find updated group")
	}

	if err := s.updateGroupCacheByUserId(joiningGroup); err != nil {
		s.log.Named("Join").Error("updateGroupCacheByUserId: joiningGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update joiningGroup cache")
	}

	groupRPC := ModelToProto(joiningGroup)

	return &proto.JoinGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) updateGroupCacheByUserId(group *model.Group) error {
	groupRPC := ModelToProto(group)
	for _, member := range group.Members {
		if err := s.cache.SetValue(groupByUserIdKey(member.ID.String()), groupRPC, s.conf.CacheTTL); err != nil {
			return fmt.Errorf("failed to update group cache: %w", err)
		}
	}

	return nil
}

func groupByUserIdKey(key string) string {
	return fmt.Sprintf("groupByUserId:%s", key)
}

func groupByTokenKey(key string) string {
	return fmt.Sprintf("groupByToken:%s", key)
}

func (s *serviceImpl) checkGroup(group *model.Group) error {
	if group.Token == "" {
		return fmt.Errorf("group token is empty")
	}
	if group.LeaderID == nil {
		return fmt.Errorf("group leader id is nil")
	}
	if len(group.Members) == 0 {
		return fmt.Errorf("group has no members")
	}
	if len(group.Members) > s.conf.Capacity {
		return fmt.Errorf("group has more than %v members (capacity exceeded)", s.conf.Capacity)
	}

	return nil
}
