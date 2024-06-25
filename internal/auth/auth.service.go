package auth

import (
	"context"
	"net/url"
	"strings"

	"github.com/isd-sgcu/rpkm67-auth/constant"
	"github.com/isd-sgcu/rpkm67-auth/internal/oauth"
	"github.com/isd-sgcu/rpkm67-auth/internal/token"
	"github.com/isd-sgcu/rpkm67-auth/internal/user"
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/auth/auth/v1"
	userProto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/auth/user/v1"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	proto.AuthServiceServer
}

type serviceImpl struct {
	proto.UnimplementedAuthServiceServer
	oauthConfig *oauth2.Config
	oauthClient oauth.GoogleOauthClient
	userSvc     user.Service
	tokenSvc    token.Service
	utils       AuthUtils
	log         *zap.Logger
}

func NewService(oauthConfig *oauth2.Config, oauthClient oauth.GoogleOauthClient, userSvc user.Service, tokenSvc token.Service, utils AuthUtils, log *zap.Logger) Service {
	return &serviceImpl{
		oauthConfig: oauthConfig,
		oauthClient: oauthClient,
		userSvc:     userSvc,
		tokenSvc:    tokenSvc,
		utils:       utils,
		log:         log,
	}
}

func (s *serviceImpl) Validate(_ context.Context, in *proto.ValidateRequest) (res *proto.ValidateResponse, err error) {
	return nil, nil
}

func (s *serviceImpl) RefreshToken(_ context.Context, in *proto.RefreshTokenRequest) (res *proto.RefreshTokenResponse, err error) {
	return nil, nil
}

func (s *serviceImpl) GetGoogleLoginUrl(_ context.Context, in *proto.GetGoogleLoginUrlRequest) (res *proto.GetGoogleLoginUrlResponse, err error) {
	URL, err := url.Parse(s.oauthConfig.Endpoint.AuthURL)
	if err != nil {
		s.log.Named("GetGoogleLoginUrl").Error("Parse: ", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	parameters := url.Values{}
	parameters.Add("client_id", s.oauthConfig.ClientID)
	parameters.Add("scope", strings.Join(s.oauthConfig.Scopes, " "))
	parameters.Add("redirect_uri", s.oauthConfig.RedirectURL)
	parameters.Add("response_type", "code")
	URL.RawQuery = parameters.Encode()
	url := URL.String()

	return &proto.GetGoogleLoginUrlResponse{
		Url: url,
	}, nil
}

func (s *serviceImpl) VerifyGoogleLogin(_ context.Context, in *proto.VerifyGoogleLoginRequest) (res *proto.VerifyGoogleLoginResponse, err error) {
	code := in.Code
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "No code is provided")
	}

	email, err := s.oauthClient.GetUserEmail(code)
	if err != nil {
		s.log.Named("VerifyGoogleLogin").Error("GetUserEmail: ", zap.Error(err))
		switch err.Error() {
		case "Invalid code":
			return nil, status.Error(codes.InvalidArgument, "Invalid code")
		default:
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}

	user, err := s.userSvc.FindByEmail(context.Background(), &userProto.FindByEmailRequest{Email: email})
	userId := user.User.Id
	role := user.User.Role
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			s.log.Named("VerifyGoogleLogin").Error("FindByEmail: ", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		switch st.Code() {
		case codes.NotFound:
			role := "user"
			if s.utils.IsStudentIdInMap(email) {
				role = "staff"
			}

			createUser := &userProto.CreateUserRequest{
				Email: email,
				Role:  role,
			}

			createdUser, err := s.userSvc.Create(context.Background(), createUser)
			if err != nil {
				s.log.Named("VerifyGoogleLogin").Error("Create: ", zap.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			userId = createdUser.User.Id
			role = createdUser.User.Role

		default:
			s.log.Named("VerifyGoogleLogin").Error("FindByEmail: ", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	credentials, err := s.tokenSvc.GetCredentials(userId, constant.Role(role))
	if err != nil {
		s.log.Named("VerifyGoogleLogin").Error("GetCredentials: ", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.VerifyGoogleLoginResponse{
		Credential: &proto.Credential{
			AccessToken:  credentials.AccessToken,
			RefreshToken: credentials.RefreshToken,
			ExpiresIn:    int32(credentials.ExpiresIn),
		},
	}, nil

}
