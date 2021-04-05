package background

import (
	"errors"
	m "github.com/chiefsend/api/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
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
	// test client
	r := asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URI")}
	client := asynq.NewClient(r)

	task := NewDeleteShareTaks(share)
	// Process the task immediately.
	_, err := client.Enqueue(task)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	var sh m.Share
	err = db.Where("ID = ?", share.ID.String()).First(&sh).Error
	assert.False(t, errors.Is(err, gorm.ErrRecordNotFound))
}
