package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

func (s *serviceImpl) FindOne(_ context.Context, in *proto.FindOneGroupRequest) (*proto.FindOneGroupResponse, error) {
	cacheKey := groupKey(in.Id)
	var cachedGroup proto.Group

	err := s.cache.GetValue(cacheKey, &cachedGroup)
	if err == nil {
		s.log.Named("FindOne").Info("GetValue: Group found in cache", zap.String("id", in.Id))
		return &proto.FindOneGroupResponse{Group: &cachedGroup}, nil
	}

	// If not found in cache, fetch from database
	group, err := s.repo.FindOne(in.Id)
	if err != nil {
		s.log.Named("FindOne").Error("FindOne: ", zap.String("id", in.Id), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	groupRPC := ModelToProto(group)
	if err := s.cache.SetValue(cacheKey, groupRPC, 3600); err != nil {
		s.log.Named("FindOne").Warn("SetValue: ", zap.String("user_id", in.Id), zap.Error(err))
	}

	return &proto.FindOneGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) FindByToken(_ context.Context, in *proto.FindByTokenGroupRequest) (*proto.FindByTokenGroupResponse, error) {
	cacheKey := groupTokenKey(in.Token)
	var cachedGroup proto.FindByTokenGroupResponse

	err := s.cache.GetValue(cacheKey, &cachedGroup)
	if err == nil {
		s.log.Named("FindByToken").Info("GetValue: Group found in cache", zap.String("token", in.Token))
		return &cachedGroup, nil
	}

	// If not found in cache, find group in database
	group, err := s.repo.FindByToken(in.Token)
	if err != nil {
		s.log.Named("FindByToken").Error("FindByToken: ", zap.String("token", in.Token), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group by token")
	}

	if err := s.checkGroup(group); err != nil {
		s.log.Named("FindByToken").Error("checkGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "group failed validation")
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

	if err := s.cache.SetValue(cacheKey, &res, 3600); err != nil {
		s.log.Named("FindByToken").Warn("SetValue: ", zap.String("token", in.Token), zap.Error(err))
	}

	return &res, nil
}

func (s *serviceImpl) UpdateConfirm(_ context.Context, in *proto.UpdateConfirmGroupRequest) (*proto.UpdateConfirmGroupResponse, error) {
	group, err := s.repo.FindByUserId(in.LeaderId)
	if err != nil {
		s.log.Named("Update").Error("FindOne group: ", zap.String("leader_id", in.LeaderId), zap.Error(err))
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

	if err := s.repo.Update(in.LeaderId, &model.Group{IsConfirmed: in.IsConfirmed}); err != nil {
		s.log.Named("Update").Error("Update: ", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update group")
	}

	updatedGroup, err := s.repo.FindOne(&userId)
	if err != nil {
		s.log.Named("Update").Error("FindOne updatedGroup: ", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find updated group")
	}

	s.updateGroupCache(updatedGroup)

	groupRPC := ModelToProto(updatedGroup)

	return &proto.UpdateConfirmGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) DeleteMember(_ context.Context, in *proto.DeleteMemberGroupRequest) (*proto.DeleteMemberGroupResponse, error) {
	group, err := s.repo.FindOne(in.UserId)
	if err != nil {
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

	var newGroup *model.Group
	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		createdGroup, err := s.repo.CreateNewGroupWithTX(tx, &userUUID)
		if err != nil {
			s.log.Named("DeleteMember").Error("CreateNewGroupWithTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.repo.MoveUserToNewGroup(tx, userUUID, createdGroup.ID); err != nil {
			s.log.Named("DeleteMember").Error("DeleteMemberFromGroupWithTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to delete member from group: %w", err)
		}

		// find created group for updating cache of poor user
		newGroup, err = s.repo.FindOne(&userUUID)
		if err != nil {
			s.log.Named("DeleteMember").Error("FindOne: ", zap.String("user_id", in.UserId), zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		s.log.Named("DeleteMember").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	updatedGroup, err := s.repo.FindOne(&leaderUUID)
	if err != nil {
		s.log.Named("DeleteMember").Error("FindOne updatedGroup: ", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find updated group")
	}

	s.updateGroupCache(updatedGroup)
	s.updateGroupCache(newGroup)

	groupRPC := ModelToProto(updatedGroup)

	return &proto.DeleteMemberGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Leave(_ context.Context, in *proto.LeaveGroupRequest) (*proto.LeaveGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		s.log.Named("Leave").Error("Parse: ", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "userId is invalid UUID")
	}

	group, err := s.repo.FindOne(&userUUID)
	if err != nil {
		s.log.Named("Leave").Error("FindOne group: ", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	if &userUUID == group.LeaderID {
		s.log.Named("Leave").Error("User is the leader of the group", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "You are the group leader, so you cannot leave")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		createdGroup, err := s.repo.CreateNewGroupWithTX(tx, &userUUID)
		if err != nil {
			s.log.Named("Leave").Error("CreateNewGroupWithTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.repo.MoveUserToNewGroup(tx, userUUID, createdGroup.ID); err != nil {
			s.log.Named("Leave").Error("DeleteMemberFromGroupWithTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to delete member from group: %w", err)
		}

		return nil
	})

	if err != nil {
		s.log.Named("Leave").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	leaderExistedGroupUUID := group.LeaderID
	existedGroup, err := s.repo.FindOne(leaderExistedGroupUUID)
	if err != nil {
		s.log.Named("Leave").Error("FindOne existedGroup: ", zap.Any("leader_id", leaderExistedGroupUUID), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	updatedGroup, err := s.repo.FindOne(&userUUID)
	if err != nil {
		s.log.Named("Leave").Error("FindOne updatedGroup: ", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find updated group")
	}

	s.updateGroupCache(existedGroup)
	s.updateGroupCache(updatedGroup)

	groupRPC := ModelToProto(updatedGroup)

	return &proto.LeaveGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Join(_ context.Context, in *proto.JoinGroupRequest) (*proto.JoinGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		s.log.Named("Join").Error("Parse: ", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "userId is invalid UUID")
	}

	existedGroup, err := s.repo.FindOne(&userUUID)
	if err != nil {
		s.log.Named("Join").Error("FindOne existedGroup: ", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find existing group of user")
	}

	if &userUUID == existedGroup.LeaderID && len(existedGroup.Members) > 1 {
		s.log.Named("Join").Error("User is the leader of a group with >1 members", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "You are the group leader, so you must kick all other members before joining another group")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		group, tokenErr := s.repo.FindByToken(in.Token)
		if tokenErr != nil {
			s.log.Named("Join").Error("FindByToken: ", zap.String("token", in.Token), zap.Error(tokenErr))
			return fmt.Errorf("failed to find group by token: %w", tokenErr)
		}

		err := s.repo.JoinGroupWithTX(tx, userUUID, group.ID)
		if err != nil {
			s.log.Named("Join").Error("JoinGroupWithTX: ", zap.String("user_id", in.UserId), zap.Error(err))
			return fmt.Errorf("failed to join group: %w", err)
		}

		if &userUUID == existedGroup.LeaderID && len(existedGroup.Members) == 1 {
			err := s.repo.DeleteGroup(tx, existedGroup.ID)
			if err != nil {
				s.log.Named("Join").Error("DeleteGroup: ", zap.String("group_id", existedGroup.ID.String()), zap.Error(err))
				return fmt.Errorf("failed to delete old group: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		s.log.Named("Join").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	updatedGroup, err := s.repo.FindOne(&userUUID)
	if err != nil {
		s.log.Named("Join").Error("FindOne updatedGroup: ", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find updated group")
	}

	s.updateGroupCache(updatedGroup)

	groupRPC := ModelToProto(updatedGroup)

	return &proto.JoinGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) updateGroupCache(group *model.Group) {
	groupRPC := ModelToProto(group)
	for _, member := range group.Members {
		if err := s.cache.SetValue(groupKey(member.ID.String()), groupRPC, 3600); err != nil {
			s.log.Named("UpdateGroupCache").Warn("SetValue: Failed to update group cache", zap.String("user_id", member.ID.String()), zap.Error(err))
		}
	}
}

func (s *serviceImpl) checkGroup(group *model.Group) error {
	if group.Token == "" {
		s.log.Named("checkGroup").Error("group token is empty")
		return fmt.Errorf("group token is empty")
	}
	if group.LeaderID == nil {
		s.log.Named("checkGroup").Error("group leader id is nil")
		return fmt.Errorf("group leader id is nil")
	}
	if len(group.Members) == 0 {
		s.log.Named("checkGroup").Error("group has no members")
		return fmt.Errorf("group has no members")
	}
	if len(group.Members) > s.conf.Capacity {
		s.log.Named("checkGroup").Error("group has more than max group size")
		return fmt.Errorf("group has more than %v members (capacity exceeded)", s.conf.Capacity)
	}

	return nil
}

func groupKey(key string) string {
	return fmt.Sprintf("group_id:%s", key)
}

func groupTokenKey(key string) string {
	return fmt.Sprintf("group_token:%s", key)
}
