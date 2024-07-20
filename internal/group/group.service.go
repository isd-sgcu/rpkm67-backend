package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-backend/config"
	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	"github.com/isd-sgcu/rpkm67-backend/internal/user"
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
	repo     Repository
	userRepo user.Repository
	cache    cache.Repository
	conf     *config.GroupConfig
	log      *zap.Logger
}

func NewService(repo Repository, userRepo user.Repository, cache cache.Repository, conf *config.GroupConfig, log *zap.Logger) Service {
	return &serviceImpl{
		repo:     repo,
		userRepo: userRepo,
		cache:    cache,
		conf:     conf,
		log:      log,
	}
}

func (s *serviceImpl) FindByUserId(_ context.Context, in *proto.FindByUserIdGroupRequest) (*proto.FindByUserIdGroupResponse, error) {
	group, err := s.findByUserId(in.UserId)
	if err != nil {
		s.log.Named("FindByUserId").Error("findByUserId: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	groupRPC := ModelToProto(group)

	return &proto.FindByUserIdGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) findByUserId(userId string) (*model.Group, error) {
	cacheKey := groupByUserIdKey(userId)
	cachedGroup := &model.Group{}

	if err := s.cache.GetValue(cacheKey, cachedGroup); err == nil {
		s.log.Named("findByUserId").Info("GetValue: Group found in cache", zap.String("userId", userId))
		return cachedGroup, nil
	}

	group, err := s.findByUserIdNoCache(userId)
	if err != nil {
		s.log.Named("findByUserId").Error("findByUserIdNoCache group: ", zap.Error(err))
		return nil, err
	}

	if err := s.cache.SetValue(cacheKey, group, s.conf.CacheTTL); err != nil {
		s.log.Named("findByUserId").Error("SetValue: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to cache group")
	}

	return group, nil
}

func (s *serviceImpl) findByUserIdNoCache(userId string) (*model.Group, error) {
	user := &model.User{}
	if err := s.userRepo.FindOne(userId, user); err != nil {
		s.log.Named("findByUserIdNoCache").Error("FindOne user: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	if user.GroupID == nil {
		err := s.repo.WithTransaction(func(tx *gorm.DB) error {
			createGroup := &model.Group{
				LeaderID: &user.ID,
			}

			if err := s.repo.CreateTX(tx, createGroup); err != nil {
				s.log.Named("findByUserIdNoCache").Error("CreateTX: ", zap.Error(err))
				return fmt.Errorf("failed to create new group: %w", err)
			}

			if err := s.userRepo.AssignGroupTX(tx, user.ID.String(), &createGroup.ID); err != nil {
				s.log.Named("findByUserIdNoCache").Error("AssignGroupTX: ", zap.Error(err))
				return fmt.Errorf("failed to assign user to group: %w", err)
			}

			return nil
		})

		if err != nil {
			s.log.Named("findByUserIdNoCache").Error("WithTransaction: ", zap.Error(err))
			return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
		}
	}

	group := &model.Group{}
	if err := s.repo.FindOne(user.GroupID.String(), group); err != nil {
		s.log.Named("findByUserIdNoCache").Error("FindOne: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find group")
	}

	return group, nil
}

func (s *serviceImpl) FindByToken(_ context.Context, in *proto.FindByTokenGroupRequest) (*proto.FindByTokenGroupResponse, error) {
	group := &model.Group{}
	if err := s.repo.FindByToken(in.Token, group); err != nil {
		s.log.Named("FindByToken").Error("FindByToken: ", zap.Error(err))
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

	return &res, nil
}

func (s *serviceImpl) UpdateConfirm(_ context.Context, in *proto.UpdateConfirmGroupRequest) (*proto.UpdateConfirmGroupResponse, error) {
	group, err := s.findByUserId(in.LeaderId)
	if err != nil {
		s.log.Named("UpdateConfirm").Error("findByUserId: ", zap.Error(err))
		return nil, err
	}

	if err := s.checkGroup(group); err != nil {
		s.log.Named("UpdateConfirm").Error("checkGroup: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "group failed validation")
	}

	if group.LeaderID.String() != in.LeaderId {
		s.log.Named("UpdateConfirm").Error("Requested leader_id is not leader of this group", zap.String("leader_id", in.LeaderId))
		return nil, status.Error(codes.PermissionDenied, "requested leader_id is not leader of this group")
	}

	group.IsConfirmed = in.IsConfirmed
	if err := s.repo.UpdateConfirm(group.ID.String(), group); err != nil {
		s.log.Named("UpdateConfirm").Error("Update: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update group")
	}

	if err := s.updateGroupCache(group); err != nil {
		s.log.Named("UpdateConfirm").Error("updateGroupCacheByUserId: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update group cache")
	}
	groupRPC := ModelToProto(group)

	return &proto.UpdateConfirmGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) DeleteMember(_ context.Context, in *proto.DeleteMemberGroupRequest) (*proto.DeleteMemberGroupResponse, error) {
	if in.LeaderId == in.UserId {
		s.log.Named("DeleteMember").Error("User is the leader of the group", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "You are the group leader, so you cannot delete yourself")
	}

	group, err := s.findByUserId(in.LeaderId)
	if err != nil {
		s.log.Named("DeleteMember").Error("findByUserId group: ", zap.Error(err))
		return nil, err
	}

	if in.LeaderId != group.LeaderID.String() {
		s.log.Named("DeleteMember").Error("Requested leader_id is not leader of this group", zap.String("leader_id", in.LeaderId))
		return nil, status.Error(codes.PermissionDenied, "requested leader_id is not leader of this group")
	}

	var found bool
	var deletedUserId *uuid.UUID
	for _, member := range group.Members {
		if member.ID.String() == in.UserId {
			found = true
			deletedUserId = &member.ID
			break
		}
	}
	if !found {
		s.log.Named("DeleteMember").Error("User is not in the group", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.NotFound, "user_id to be deleted is not in the group")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		createGroup := &model.Group{
			LeaderID: deletedUserId,
		}

		if err := s.repo.CreateTX(tx, createGroup); err != nil {
			s.log.Named("DeleteMember").Error("CreateTX: ", zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.userRepo.AssignGroupTX(tx, in.UserId, &createGroup.ID); err != nil {
			s.log.Named("DeleteMember").Error("AssignGroupTX: ", zap.Error(err))
			return fmt.Errorf("failed to assign user to new group: %w", err)
		}

		return nil
	})

	if err != nil {
		s.log.Named("DeleteMember").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	newGroup, err := s.findByUserIdNoCache(in.UserId)
	if err != nil {
		s.log.Named("DeleteMember").Error("findByUserIdNoCache newGroup: ", zap.Error(err))
		return nil, err
	}
	updatedGroup, err := s.findByUserIdNoCache(in.LeaderId)
	if err != nil {
		s.log.Named("DeleteMember").Error("findByUserIdNoCache updatedGroup: ", zap.Error(err))
		return nil, err
	}

	if err := s.updateGroupCache(newGroup); err != nil {
		s.log.Named("DeleteMember").Error("updateGroupCacheByUserId: newGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update newGroup cache")
	}
	if err := s.updateGroupCache(updatedGroup); err != nil {
		s.log.Named("DeleteMember").Error("updateGroupCacheByUserId: updatedGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update updatedGroup cache")
	}

	groupRPC := ModelToProto(updatedGroup)

	return &proto.DeleteMemberGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Leave(_ context.Context, in *proto.LeaveGroupRequest) (*proto.LeaveGroupResponse, error) {
	group, err := s.findByUserId(in.UserId)
	if err != nil {
		s.log.Named("Leave").Error("findByUserId group: ", zap.Error(err))
		return nil, err
	}

	if group.IsConfirmed {
		s.log.Named("Leave").Error("Group is confirmed", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "Group is confirmed, so you cannot leave")
	}

	if in.UserId == group.LeaderID.String() {
		s.log.Named("Leave").Error("User is the leader of the group", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "You are the group leader, so you cannot leave")
	}

	if len(group.Members) == 1 {
		s.log.Named("Leave").Error("Group has only one member", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "You are the only member in the group, so you cannot leave")
	}

	userId, err := uuid.Parse(in.UserId)
	if err != nil {
		s.log.Named("Leave").Error("Parse userId: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to parse user id")
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {
		createGroup := &model.Group{
			LeaderID: &userId,
		}

		if err := s.repo.CreateTX(tx, createGroup); err != nil {
			s.log.Named("Leave").Error("CreateTX: ", zap.Error(err))
			return fmt.Errorf("failed to create new group: %w", err)
		}

		if err := s.userRepo.AssignGroupTX(tx, in.UserId, &createGroup.ID); err != nil {
			s.log.Named("Leave").Error("AssignGroupTX: ", zap.Error(err))
			return fmt.Errorf("failed to assign user to new group: %w", err)
		}

		return nil
	})

	if err != nil {
		s.log.Named("Leave").Error("WithTransaction: ", zap.Error(err))
		return nil, status.Error(codes.Internal, fmt.Sprintf("transaction failed: %s", err.Error()))
	}

	newGroup, err := s.findByUserIdNoCache(in.UserId)
	if err != nil {
		s.log.Named("Leave").Error("findByUserIdNoCache group: ", zap.Error(err))
		return nil, err
	}
	updatedGroup, err := s.findByUserIdNoCache(group.LeaderID.String())
	if err != nil {
		s.log.Named("Leave").Error("findByUserIdNoCache group: ", zap.Error(err))
		return nil, err
	}

	if err := s.updateGroupCache(newGroup); err != nil {
		s.log.Named("Leave").Error("updateGroupCacheByUserId: newGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update newGroup cache")
	}
	if err := s.updateGroupCache(updatedGroup); err != nil {
		s.log.Named("Leave").Error("updateGroupCacheByUserId: updatedGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update updatedGroup cache")
	}

	groupRPC := ModelToProto(updatedGroup)

	return &proto.LeaveGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) Join(_ context.Context, in *proto.JoinGroupRequest) (*proto.JoinGroupResponse, error) {
	group, err := s.findByUserId(in.UserId)
	if err != nil {
		s.log.Named("Join").Error("findByUserId group: ", zap.Error(err))
		return nil, err
	}

	if group.IsConfirmed {
		s.log.Named("Join").Error("Group is confirmed", zap.String("user_id", in.UserId))
		return nil, status.Error(codes.PermissionDenied, "Group is confirmed, so you cannot leave to join other groups")
	}

	joiningGroup := &model.Group{}
	if err := s.repo.FindByToken(in.Token, joiningGroup); err != nil {
		s.log.Named("Join").Error("FindByToken joiningGroup TX: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to find joining group by token")
	}

	for _, member := range joiningGroup.Members {
		if member.ID.String() == in.UserId {
			s.log.Named("Join").Error("User is already in the group", zap.String("user_id", in.UserId))
			return nil, status.Error(codes.PermissionDenied, "user is already in the group")
		}
	}

	if len(joiningGroup.Members) >= s.conf.Capacity {
		s.log.Named("Join").Error("Group is full", zap.String("token", in.Token))
		return nil, status.Error(codes.PermissionDenied, "group is full")
	}

	prevGroup, err := s.findByUserId(in.UserId)
	if err != nil {
		s.log.Named("Join").Error("findByUserId prevGroup: ", zap.Error(err))
		return nil, err
	}

	err = s.repo.WithTransaction(func(tx *gorm.DB) error {

		if err := s.userRepo.AssignGroupTX(tx, in.UserId, &joiningGroup.ID); err != nil {
			s.log.Named("findByUserId").Error("AssignGroupTX: ", zap.Error(err))
			return fmt.Errorf("failed to assign user to group: %w", err)
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

	joiningGroup.Members = append(joiningGroup.Members, prevGroup.Members[0])

	if err := s.updateGroupCache(joiningGroup); err != nil {
		s.log.Named("Join").Error("updateGroupCacheByUserId: joiningGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update joiningGroup cache")
	}

	groupRPC := ModelToProto(joiningGroup)

	return &proto.JoinGroupResponse{Group: groupRPC}, nil
}

func (s *serviceImpl) updateGroupCache(group *model.Group) error {
	for _, member := range group.Members {
		if err := s.cache.SetValue(groupByUserIdKey(member.ID.String()), group, s.conf.CacheTTL); err != nil {
			return fmt.Errorf("failed to update group cache by user id: %w", err)
		}
	}

	if err := s.cache.SetValue(groupByTokenKey(group.Token), group, s.conf.CacheTTL); err != nil {
		return fmt.Errorf("failed to update group cache by token: %w", err)
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
