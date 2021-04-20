package background

import (
	"github.com/chiefsend/api/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"log"
	"os"
	"testing"
	"time"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	dab, err := models.GetDatabase()
	if err != nil {
		log.Fatal("database brok")
	}
	db = dab
	_ = db.AutoMigrate(&models.Share{})
	_ = db.AutoMigrate(&models.Attachment{})

	os.Exit(m.Run())
}

func TestHandleDeleteShareTask(t *testing.T) {
	var share = models.Share{
		ID:            uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
		IsPublic:      false,
		IsTemporary:   false,
	}
	db.Create(&share)
	defer db.Delete(&share)
	go StartBackgroundWorkers()
	defer StopBackgroundWorkers()

	t.Run("happy path", func(t *testing.T) {
		task := NewDeleteShareTask(share)
		err := EnqueueJob(task, nil)
		assert.Nil(t, err)
		time.Sleep(time.Second)
		// check
		var sh models.Share
		err = db.Where("ID = ?", share.ID.String()).First(&sh).Error
		assert.Error(t, gorm.ErrRecordNotFound, err)
	})
}

func TestHandleContinuousDeleteTask(t *testing.T) {
	var shares = []models.Share {
		{
			ID: uuid.MustParse("9788fedd-d840-4ad5-9824-05fa3d59b686"),
			CreatedAt: time.Now().Add(-30*time.Hour), // should be deleted
			IsTemporary: true,
		},
		{
			ID: uuid.MustParse("b8b4d8f2-0a58-4400-ad41-a6f39b82e9da"),
			CreatedAt: time.Now().Add(-10*time.Hour), // should not be deleted
			IsTemporary: true,
		},
	}
	for _, sh := range shares {
		db.Create(&sh)
		defer db.Delete(&sh)
	}

	t.Run("happy path", func(t *testing.T) {
		task := NewContinuousDeleteTask()
		err := EnqueueJob(task, nil)
		assert.Nil(t, err)
		time.Sleep(time.Second)
		// assertions
		var actual []models.Share
		err = db.Find(&actual).Error
		assert.Nil(t, err)
		assert.Len(t, actual, 1)
	})
}