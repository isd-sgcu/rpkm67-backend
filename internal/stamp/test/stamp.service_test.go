package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StampServiceTest struct {
	suite.Suite
	// controller *gomock.Controller
	// logger     *zap.Logger
}

func TestStampService(t *testing.T) {
	suite.Run(t, new(StampServiceTest))
}

func (t *StampServiceTest) SetupTest() {}

func (t *StampServiceTest) TestFindByUserIdSuccess() {
}

func (t *StampServiceTest) TestStampByUserIdSuccess() {
}
