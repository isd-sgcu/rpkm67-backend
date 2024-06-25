package user

import (
	"context"
	"errors"

	"github.com/isd-sgcu/rpkm67-auth/constant"
	"github.com/isd-sgcu/rpkm67-auth/internal/model"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/auth/user/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Service interface {
	proto.UserServiceServer
}

type serviceImpl struct {
	proto.UnimplementedUserServiceServer
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) proto.UserServiceServer {
	return &serviceImpl{
		repo: repo,
		log:  log,
	}
}

func (s *serviceImpl) Create(_ context.Context, req *proto.CreateUserRequest) (res *proto.CreateUserResponse, err error) {
	createUser := &model.User{
		Email: req.Email,
		Role:  constant.Role(req.Role),
	}

	err = s.repo.Create(createUser)
	if err != nil {
		s.log.Named("Create").Error("Create: ", zap.Error(err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, status.Error(codes.AlreadyExists, constant.DuplicateEmailErrorMessage)
		}
		return nil, err
	}

	return &proto.CreateUserResponse{
		User: ModelToProto(createUser),
	}, nil
}

func (s *serviceImpl) FindOne(_ context.Context, req *proto.FindOneUserRequest) (res *proto.FindOneUserResponse, err error) {
	user := &model.User{}

	err = s.repo.FindOne(req.Id, user)
	if err != nil {
		s.log.Named("FindOne").Error("FindOne: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, constant.UserNotFoundErrorMessage)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.FindOneUserResponse{
		User: ModelToProto(user),
	}, nil
}

func (s *serviceImpl) FindByEmail(_ context.Context, req *proto.FindByEmailRequest) (res *proto.FindByEmailResponse, err error) {
	user := &model.User{}

	err = s.repo.FindByEmail(req.Email, user)
	if err != nil {
		s.log.Named("FindByEmail").Error("FindByEmail: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, constant.UserNotFoundErrorMessage)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.FindByEmailResponse{
		User: ModelToProto(user),
	}, nil
}

func ModelToProto(in *model.User) *proto.User {
	return &proto.User{
		Id:        in.ID.String(),
		Email:     in.Email,
		Firstname: in.Firstname,
		Lastname:  in.Lastname,
		Role:      in.Role.String(),
	}
}
