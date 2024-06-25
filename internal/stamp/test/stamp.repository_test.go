package test

import (
	"fmt"
	"testing"

	"github.com/isd-sgcu/rpkm67-auth/internal/model"
	"github.com/isd-sgcu/rpkm67-auth/internal/stamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type StampRepositoryTest struct {
	suite.Suite
	db   *gorm.DB
	repo stamp.Repository
}

func TestStampRepository(t *testing.T) {
	suite.Run(t, new(StampRepositoryTest))
}

func (t *StampRepositoryTest) SetupTest() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", "localhost", "5433", "root", "1234", "rpkm67_test_db", "")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})

	assert.NoError(t.T(), err)

	_ = db.Migrator().DropTable(&model.Stamp{})

	err = db.AutoMigrate(&model.User{}, &model.Stamp{})
	assert.NoError(t.T(), err)

	t.repo = stamp.NewRepository(db)
	t.db = db
}

func (t *StampRepositoryTest) TestFindByUserIdSuccess() {
	// createStamp := &model.Stamp{
	// 	PointA: 1,
	// 	PointB: 2,
	// 	PointC: 3,
	// 	PointD: 4,
	// 	Stamp:  "01010101010",
	// }

	// err := t.repo.FindByUserId(cre)
	// assert.Nil(t.T(), err)
}
