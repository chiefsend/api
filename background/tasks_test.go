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

var share = models.Share{
	ID:            uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
	IsPublic:      false,
	IsTemporary:   false,
}

var db *gorm.DB

func TestMain(m *testing.M) {
	dab, err := models.GetDatabase()
	if err != nil {
		log.Fatal("database brok")
	}
	db = dab
	_ = db.AutoMigrate(&models.Share{})
	_ = db.AutoMigrate(&models.Attachment{})

	code := m.Run()
	os.Exit(code)
}

func TestDeleteShareTask(t *testing.T) {
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
