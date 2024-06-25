package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/isd-sgcu/rpkm67-auth/internal/dto"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type GoogleOauthClient interface {
	GetUserEmail(code string) (string, error)
}

type googleOauthClientImpl struct {
	oauthConfig *oauth2.Config
	log         *zap.Logger
}

func NewGoogleOauthClient(oauthConfig *oauth2.Config, log *zap.Logger) GoogleOauthClient {
	return &googleOauthClientImpl{
		oauthConfig,
		log,
	}
}

var (
	InvalidCode   = errors.New("Invalid code")
	HttpError     = errors.New("Unable to get user info")
	IOError       = errors.New("Unable to read google response")
	InvalidFormat = errors.New("Google sent unexpected format")
)

func (c *googleOauthClientImpl) GetUserEmail(code string) (string, error) {
	token, err := c.oauthConfig.Exchange(context.TODO(), code)
	if err != nil {
		c.log.Named("GetUserEmail").Error("Exchange: ", zap.Error(err))
		return "", InvalidCode
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		c.log.Named("GetUserEmail").Error("Get: ", zap.Error(err))
		return "", HttpError
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Named("GetUserEmail").Error("ReadAll: ", zap.Error(err))
		return "", IOError
	}

	var parsedResponse dto.GoogleUserEmailResponse
	if err = json.Unmarshal(response, &parsedResponse); err != nil {
		c.log.Named("GetUserEmail").Error("Unmarshal: ", zap.Error(err))
		return "", InvalidFormat
	}

	return parsedResponse.Email, nil
}
