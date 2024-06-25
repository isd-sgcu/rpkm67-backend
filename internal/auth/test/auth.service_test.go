package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthServiceTest struct {
	suite.Suite
	// controller *gomock.Controller
	// logger     *zap.Logger
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceTest))
}

func (t *AuthServiceTest) SetupTest() {}

func (t *AuthServiceTest) TestSignUpSuccess() {

}
