package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/group/v1"
	"go.uber.org/zap"
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
	// var cachedGroup proto.Group

	// // try to retreive from cache
	// err := s.cache.GetValue(cacheKey, &cachedGroup)
	// if err == nil {
	// 	s.log.Info("Group found in cache", zap.String("user_id", in.UserId))
	// 	return &proto.FindOneGroupResponse{Group: &cachedGroup}, nil
	// }

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
	cacheKey := fmt.Sprintf("token:%s", in.Token)

	// var cachedGroup proto.Group

	// // try to retreive from cache
	// err := s.cache.GetValue(cacheKey, &cachedGroup)
	// if err == nil {
	// 	s.log.Info("Group found in cache", zap.String("token", in.Token))
	// 	return &proto.FindByTokenGroupResponse{Leader: &cachedGroup}, nil
	// }

	// if not found cache, find group in database
	group, err := s.repo.FindByToken(in.Token)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("token", in.Token), zap.Error(err))
		return nil, err
	}

	userInfo := make([]*proto.UserInfo, 0, len(group.Members))
	var LeaderInfo proto.UserInfo
	for _, m := range group.Members {
		user := proto.UserInfo{
			Id:        m.ID.String(),
			Firstname: m.Firstname,
			Lastname:  m.Lastname,
			ImageUrl:  m.PhotoUrl,
		}
		userInfo = append(userInfo, &user)
		if user.Id == group.LeaderID {
			LeaderInfo = proto.UserInfo{
				Id:        m.ID.String(),
				Firstname: m.Firstname,
				Lastname:  m.Lastname,
				ImageUrl:  m.PhotoUrl,
			}
		}
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
		s.log.Warn("Failed to set group in cache", zap.String("token", in.Token), zap.Error(err))
	}

	res := proto.FindByTokenGroupResponse{
		Id:     group.ID.String(),
		Leader: &LeaderInfo,
		Token:  group.Token,
	}

	s.log.Info("FindByToken group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("token", in.Token),
		zap.Int("member_count", len(userInfo)),
		zap.Bool("from_cache", false))

	return &res, nil
}

func (s *serviceImpl) Update(ctx context.Context, in *proto.UpdateGroupRequest) (*proto.UpdateGroupResponse, error) {
	userUUID, err := uuid.Parse(in.LeaderId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, err
	}

	if err := s.repo.Update(userUUID, group); err != nil {
		s.log.Error("Failed to update group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, err
	}

	cacheKey := fmt.Sprintf("group:%s", in.LeaderId)
	if err := s.cache.SetValue(cacheKey, group, 3600); err != nil { // cache นาน 1 ชั่วโมง
		s.log.Warn("Failed to set group in cache", zap.String("leader_id", in.LeaderId), zap.Error(err))
	}

	res := proto.UpdateGroupResponse{
		Group: &proto.Group{
			Id:          group.ID.String(),
			LeaderID:    group.LeaderID,
			Token:       group.Token,
			Members:     nil,
			IsConfirmed: group.IsConfirmed,
		},
	}

	s.log.Info("UpdateGroup group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("user_id", in.LeaderId),
		zap.Bool("is_confirmed", group.IsConfirmed))

	return &res, nil
}

func (s *serviceImpl) Join(ctx context.Context, in *proto.JoinGroupRequest) (*proto.JoinGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindByToken(in.Token)
	if err != nil {
		s.log.Error("Failed to find group with Token", zap.String("token", in.Token), zap.Error(err))
	}

	for _, member := range group.Members {
		if member.ID.String() == in.UserId {
			return nil, fmt.Errorf("user %s is already a member of the group", in.UserId)
		} else if member.ID.String() == group.LeaderID {
			return nil, fmt.Errorf("user %s is the leader of the group", in.UserId)
		} else if member.ID.String() != in.UserId {
			return nil, fmt.Errorf("group is already full", in.UserId)
		}
	}

	chkUser, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find user", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, err
	} else if len(chkUser.Members) > 1 {
		return nil, fmt.Errorf("user %s is already a member of a group", in.UserId)
	}

	if err := s.repo.Join(userUUID, group); err != nil {
		s.log.Error("Failed to join group", zap.String("user_id", userUUID.String()), zap.Error(err))
		return nil, err
	}

	cacheKey := fmt.Sprintf("group:%s", in.UserId)
	if err := s.cache.SetValue(cacheKey, group, 3600); err != nil { // cache นาน 1 ชั่วโมง
		s.log.Warn("Failed to set group in cache", zap.String("user_id", in.UserId), zap.Error(err))
	}

	userInfo := make([]*proto.UserInfo, 0, len(group.Members))
	for _, m := range group.Members {
		userInfo = append(userInfo, &proto.UserInfo{
			Id:        m.ID.String(),
			Firstname: m.Firstname,
			Lastname:  m.Lastname,
			ImageUrl:  m.PhotoUrl,
		})
	}

	res := proto.JoinGroupResponse{
		Group: &proto.Group{
			Id:          group.ID.String(),
			LeaderID:    group.LeaderID,
			Token:       group.Token,
			Members:     userInfo,
			IsConfirmed: group.IsConfirmed,
		},
	}

	s.log.Info("JoinGroup group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("user_id", in.UserId))

	return &res, nil
}

func (s *serviceImpl) DeleteMember(ctx context.Context, in *proto.DeleteMemberGroupRequest) (*proto.DeleteMemberGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, err
	}

	if in.UserId == in.LeaderId {
		return nil, fmt.Errorf("user %s is the leader of the group", in.UserId)
	}

	for _, member := range group.Members {
		if member.ID.String() != in.UserId {
			return nil, fmt.Errorf("user is not in group", in.UserId)
		}
	}

	if err := s.repo.DeleteMember(userUUID, group); err != nil {
		s.log.Error("Failed to delete member", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, err
	}

	cacheKey := fmt.Sprintf("group:%s", in.UserId)
	if err := s.cache.SetValue(cacheKey, group, 3600); err != nil { // cache นาน 1 ชั่วโมง
		s.log.Warn("Failed to set group in cache", zap.String("user_id", in.UserId), zap.Error(err))
	}

	res := proto.DeleteMemberGroupResponse{
		Group: &proto.Group{
			Id:          group.ID.String(),
			LeaderID:    group.LeaderID,
			Token:       group.Token,
			Members:     nil,
			IsConfirmed: group.IsConfirmed,
		},
	}

	s.log.Info("DeleteMember group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("user_id", in.UserId))

	return &res, nil
}

func (s *serviceImpl) Leave(ctx context.Context, in *proto.LeaveGroupRequest) (*proto.LeaveGroupResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, err
	}

	if in.UserId == group.LeaderID {
		return nil, fmt.Errorf("user %s is the leader of the group", in.UserId)
	}

	if err := s.repo.DeleteMember(group, userUUID); err != nil {
		s.log.Error("Failed to leave group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, err
	}

	//create group for deleted member here

	cacheKey := fmt.Sprintf("group:%s", in.UserId)
	if err := s.cache.SetValue(cacheKey, group, 3600); err != nil { // cache นาน 1 ชั่วโมง
		s.log.Warn("Failed to set group in cache", zap.String("user_id", in.UserId), zap.Error(err))
	}

	res := proto.LeaveGroupResponse{
		Group: &proto.Group{
			Id:          group.ID.String(),
			LeaderID:    group.LeaderID,
			Token:       group.Token,
			Members:     nil,
			IsConfirmed: group.IsConfirmed,
		},
	}

	s.log.Info("LeaveGroup group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("user_id", in.UserId))

	return &res, nil
}

func (s *serviceImpl) SelectBaan(ctx context.Context, in *proto.SelectBaanRequest) (*proto.SelectBaanResponse, error) {
	userUUID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	group, err := s.repo.FindOne(userUUID)
	if err != nil {
		s.log.Error("Failed to find group", zap.String("user_id", in.UserId), zap.Error(err))
		return nil, err
	}

	for _, member := range group.Members {
		if member.ID.String() != in.UserId {
			return nil, fmt.Errorf("user is not in group", in.UserId)
		}
	}

	selection := &Selection{
		GroupID: group.ID,
		Baan:    in.BaanId,
		Order:   int(in.Order),
	}

	err = s.repo.Create(selection)
	if err != nil {
		s.log.Error("Failed to create selection", zap.Error(err))
		return nil, err
	}

	cacheKey := fmt.Sprintf("group:%s", in.GroupId)
	err = s.cache.SetValue(cacheKey, selection, 3600)
	if err != nil {
		s.log.Error("Failed to cache selection", zap.Error(err))
	}

	res := proto.SelectBaanGroupResponse{
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

func (s *serviceImpl) Create(ctx context.Context, in *proto.LeaveGroupRequest) (*proto.LeaveGroupResponse, error) {
	group := &Group{
		ID:       uuid.New(),
		LeaderID: in.LeaderId,
		Token:    in.Token,
	}

	group, err := s.repo.FindByToken(in.Token)

	if err := s.repo.Create(group); err != nil {
		s.log.Error("Failed to create group", zap.String("leader_id", in.LeaderId), zap.Error(err))
		return nil, err
	}

	cacheKey := fmt.Sprintf("group:%s", in.LeaderId)
	if err := s.cache.SetValue(cacheKey, group, 3600); err != nil { // cache นาน 1 ชั่วโมง
		s.log.Warn("Failed to set group in cache", zap.String("leader_id", in.LeaderId), zap.Error(err))
	}

	res := proto.CreateGroupResponse{
		Group: &proto.Group{
			Id:          group.ID.String(),
			LeaderID:    group.LeaderID,
			Token:       group.Token,
			Members:     nil,
			IsConfirmed: group.IsConfirmed,
		},
	}

	s.log.Info("CreateGroup group service completed",
		zap.String("group_id", group.ID.String()),
		zap.String("leader_id", in.LeaderId))

	return &res, nil
}
