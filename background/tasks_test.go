package background

import (
	"errors"
	g "github.com/chiefsend/api/globals"
	m "github.com/chiefsend/api/models"
	u "github.com/chiefsend/api/util"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestDeleteShareTaks(t *testing.T) {
	u.Reset()
	db, _ := m.GetDatabase()
	// test client
	r := asynq.RedisClientOpt{Addr: g.Conf.RedisAddr}
	client := asynq.NewClient(r)

	task := NewDeleteShareTaks(u.Shares[0])
	// Process the task immediately.
	_, err := client.Enqueue(task)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	var sh m.Share
	err = db.Where("ID = ?", u.Shares[0].ID.String()).First(&sh).Error
	assert.False(t, errors.Is(err, gorm.ErrRecordNotFound))
}
