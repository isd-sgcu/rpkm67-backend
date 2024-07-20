package count

import (
	"context"
	"fmt"

	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/count/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	proto.CountServiceServer
}

type serviceImpl struct {
	proto.UnimplementedCountServiceServer
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) Service {
	return &serviceImpl{
		repo: repo,
		log:  log,
	}
}

func (s *serviceImpl) Create(ctx context.Context, in *proto.CreateCountRequest) (*proto.CreateCountResponse, error) {
	count := model.Count{
		Name: in.Name,
	}

	err := s.repo.Create(&count)
	if err != nil {
		s.log.Named("Create").Error(fmt.Sprintf("Create: name=%s", in.Name), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := proto.CreateCountResponse{
		Count: &proto.Count{
			Name: count.Name,
		},
	}

	return &res, nil
}
