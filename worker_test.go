package main

import (
	"errors"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestDeleteShareTaks(t *testing.T) {
	Reset()
	db, err := GetDatabase()
	assert.Nil(t, err)
	// test client
	r := asynq.RedisClientOpt{Addr: config.redisAddr}
	client := asynq.NewClient(r)

	task := NewDeleteShareTaks(shares[0])
	// Process the task immediately.
	_, err = client.Enqueue(task)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	var sh Share
	err = db.Where("ID = ?", shares[0].ID.String()).First(&sh).Error
	assert.False(t, errors.Is(err, gorm.ErrRecordNotFound))
}
