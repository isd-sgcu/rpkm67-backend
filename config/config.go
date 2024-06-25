package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AppConfig struct {
	Port string
	Env  string
}

type DbConfig struct {
	Url string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type JwtConfig struct {
	Secret        string
	AccessTTL     int
	RefreshTTL    int
	Issuer        string
	ResetTokenTTL int
}

type OauthConfig struct {
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectUri  string `mapstructure:"redirect_uri"`
}

type Config struct {
	App   AppConfig
	Db    DbConfig
	Redis RedisConfig
	Jwt   JwtConfig
	Oauth OauthConfig
}

func LoadConfig() (*Config, error) {
	if os.Getenv("APP_ENV") == "" {
		err := godotenv.Load(".env")
		if err != nil {
			return nil, err
		}
	}

	appConfig := AppConfig{
		Port: os.Getenv("APP_PORT"),
		Env:  os.Getenv("APP_ENV"),
	}

	dbConfig := DbConfig{
		Url: os.Getenv("DB_URL"),
	}

	redisPort, err := strconv.ParseInt(os.Getenv("REDIS_PORT"), 10, 64)
	if err != nil {
		return nil, err
	}

	redisConfig := RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     int(redisPort),
		Password: os.Getenv("REDIS_PASSWORD"),
	}

	accessTTL, err := strconv.ParseInt(os.Getenv("JWT_ACCESS_TTL"), 10, 64)
	if err != nil {
		return nil, err
	}
	refreshTTL, err := strconv.ParseInt(os.Getenv("JWT_REFRESH_TTL"), 10, 64)
	if err != nil {
		return nil, err
	}
	resetTokenTTL, err := strconv.ParseInt(os.Getenv("JWT_RESET_TOKEN_TTL"), 10, 64)
	if err != nil {
		return nil, err
	}

	jwtConfig := JwtConfig{
		Secret:        os.Getenv("JWT_SECRET"),
		AccessTTL:     int(accessTTL),
		RefreshTTL:    int(refreshTTL),
		Issuer:        os.Getenv("JWT_ISSUER"),
		ResetTokenTTL: int(resetTokenTTL),
	}

	oauthConfig := OauthConfig{
		ClientId:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		RedirectUri:  os.Getenv("OAUTH_REDIRECT_URI"),
	}

	return &Config{
		App:   appConfig,
		Db:    dbConfig,
		Redis: redisConfig,
		Jwt:   jwtConfig,
		Oauth: oauthConfig,
	}, nil
}

func (ac *AppConfig) IsDevelopment() bool {
	return ac.Env == "development"
}

func LoadOauthConfig(oauth OauthConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     oauth.ClientId,
		ClientSecret: oauth.ClientSecret,
		RedirectURL:  oauth.RedirectUri,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	}
}
