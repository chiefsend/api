package background

import (
	"errors"
	m "github.com/chiefsend/api/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

var share = m.Share {
	ID:            uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
	Name:          "TestFinalPrivate",
	DownloadLimit: 100,
	IsPublic:      false,
	IsTemporary:   false,
	Password:      "test123",
}

func TestDeleteShareTaks(t *testing.T) {
	db, _ := m.GetDatabase()
	task := NewDeleteShareTaks(share)
	err := EnqueueJob(task, nil)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	var sh m.Share
	err = db.Where("ID = ?", share.ID.String()).First(&sh).Error
	assert.False(t, errors.Is(err, gorm.ErrRecordNotFound))
}
