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
		Id:       group.ID.String(),
		LeaderID: group.LeaderID,
		Token:    group.Token,
		Members:  userInfo,
		Baans:    nil,
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
