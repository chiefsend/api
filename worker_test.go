package main

import (
	"errors"
	g "github.com/chiefsend/api/globals"
	m "github.com/chiefsend/api/models"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestDeleteShareTaks(t *testing.T) {
	Reset()
	// test client
	r := asynq.RedisClientOpt{Addr: g.Conf.RedisAddr}
	client := asynq.NewClient(r)

	task := NewDeleteShareTaks(shares[0])
	// Process the task immediately.
	_, err := client.Enqueue(task)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	var sh m.Share
	err = g.Db.Where("ID = ?", shares[0].ID.String()).First(&sh).Error
	assert.False(t, errors.Is(err, gorm.ErrRecordNotFound))
}
