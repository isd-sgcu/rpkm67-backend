package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PinServiceTest struct {
	suite.Suite
	// controller *gomock.Controller
	// logger     *zap.Logger
}

func TestPinService(t *testing.T) {
	suite.Run(t, new(PinServiceTest))
}

func (t *PinServiceTest) SetupTest() {}

func (t *PinServiceTest) TestFindAllSuccess() {
}

func (t *PinServiceTest) TestResetPinSuccess() {
}
