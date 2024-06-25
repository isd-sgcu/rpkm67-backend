package token

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-auth/config"
	"github.com/isd-sgcu/rpkm67-auth/constant"
	"github.com/isd-sgcu/rpkm67-auth/internal/cache"
	"github.com/isd-sgcu/rpkm67-auth/internal/dto"
	"github.com/isd-sgcu/rpkm67-auth/internal/jwt"
	"go.uber.org/zap"
)

type Service interface {
	GetCredentials(userId string, role constant.Role) (*dto.Credentials, error)
	CreateCredentials(userId string, role constant.Role) (*dto.Credentials, error)
	RefreshToken(refreshToken string) (*dto.Credentials, error)
	GetConfig() *config.JwtConfig
}

type serviceImpl struct {
	jwtService jwt.Service
	cache      cache.Repository
	tokenUtils TokenUtils
	log        *zap.Logger
}

func NewService(jwtService jwt.Service, cache cache.Repository, tokenUtils TokenUtils, log *zap.Logger) Service {
	return &serviceImpl{
		jwtService: jwtService,
		cache:      cache,
		tokenUtils: tokenUtils,
		log:        log,
	}
}

func (s *serviceImpl) GetCredentials(userId string, role constant.Role) (*dto.Credentials, error) {
	credentials := &dto.Credentials{}
	err := s.cache.GetValue(sessionKey(userId), credentials)
	if err != nil {
		s.log.Named("tokenSvc").Named("GetCredentials").Info("No session found in cache for user", zap.String("userId", userId))
		credentials, err = s.CreateCredentials(userId, role)
		if err != nil {
			s.log.Named("GetCredentials").Error("CreateCredentials: ", zap.Error(err))
			return nil, err
		}
	}

	_, err = s.jwtService.ValidateToken(credentials.AccessToken)
	if err != nil { // still have refreshToken but accessToken is expired
		err := s.cache.DeleteValue(sessionKey(userId))
		if err != nil {
			s.log.Named("GetCredentials").Error("DeleteValue: ", zap.Error(err))
			return nil, err
		}

		accessToken, err := s.jwtService.CreateToken(userId, role)
		if err != nil {
			s.log.Named("GetCredentials").Error("CreateToken: ", zap.Error(err))
			return nil, err
		}

		newCredentials := &dto.Credentials{
			AccessToken:  accessToken,
			RefreshToken: credentials.RefreshToken,
			ExpiresIn:    s.jwtService.GetConfig().AccessTTL,
		}

		err = s.cache.SetValue(sessionKey(userId), newCredentials, s.jwtService.GetConfig().AccessTTL)
		if err != nil {
			s.log.Named("GetCredentials").Error("SetValue: ", zap.Error(err))
			return nil, err
		}

		return newCredentials, nil
	}

	return credentials, nil
}

func (s *serviceImpl) CreateCredentials(userId string, role constant.Role) (*dto.Credentials, error) {
	accessToken, err := s.jwtService.CreateToken(userId, role)
	if err != nil {
		s.log.Named("CreateCredentials").Error("CreateToken: ", zap.Error(err))
		return nil, err
	}

	refreshToken := createRefreshToken()

	err = s.cache.SetValue(refreshKey(refreshToken), &dto.RefreshTokenCache{
		UserID: userId,
		Role:   role,
	}, s.jwtService.GetConfig().RefreshTTL)
	if err != nil {
		s.log.Named("CreateCredentials").Error("SetValue refresh: ", zap.Error(err))
		return nil, err
	}

	credentials := &dto.Credentials{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	err = s.cache.SetValue(sessionKey(userId), credentials, s.jwtService.GetConfig().AccessTTL)
	if err != nil {
		s.log.Named("CreateCredentials").Error("SetValue session: ", zap.Error(err))
		return nil, err
	}

	return credentials, nil
}

func (s *serviceImpl) RefreshToken(refreshToken string) (*dto.Credentials, error) {
	refreshCache := &dto.RefreshTokenCache{}
	err := s.cache.GetValue(refreshKey(refreshToken), refreshCache)
	if err != nil {
		s.log.Named("RefreshToken").Error("GetValue: ", zap.Error(err))
		return nil, err
	} else if (refreshCache == &dto.RefreshTokenCache{}) {
		s.log.Named("RefreshToken").Info("GetValue: refresh token not found")
		return nil, fmt.Errorf("refresh token not found")
	}

	err = s.cache.DeleteValue(refreshKey(refreshToken))
	if err != nil {
		s.log.Named("RefreshToken").Error("DeleteValue refresh: ", zap.Error(err))
		return nil, err
	}

	err = s.cache.DeleteValue(sessionKey(refreshCache.UserID))
	if err != nil {
		s.log.Named("RefreshToken").Error("DeleteValue session: ", zap.Error(err))
		return nil, err
	}

	credentials, err := s.CreateCredentials(refreshCache.UserID, refreshCache.Role)
	if err != nil {
		s.log.Named("RefreshToken").Error("CreateCredentials: ", zap.Error(err))
		return nil, err
	}

	return credentials, nil
}

func (s *serviceImpl) GetConfig() *config.JwtConfig {
	return s.jwtService.GetConfig()
}

func createRefreshToken() string {
	return uuid.New().String()
}

func refreshKey(refreshToken string) string {
	return fmt.Sprintf("refresh:%s", refreshToken)
}

func sessionKey(userId string) string {
	return fmt.Sprintf("session:%s", userId)
}
