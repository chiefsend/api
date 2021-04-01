package main

import (
	"context"
	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	DeleteShare = "share:delete"
	ShareEmail  = "email:share"
)

// Tasks
func NewShareEmailTask(share Share) *asynq.Task {
	payload := map[string]interface{}{"share_id": share.ID.String()}
	return asynq.NewTask(ShareEmail, payload)
}

func NewDeleteShareTaks(share Share) *asynq.Task {
	payload := map[string]interface{}{"share_id": share.ID.String()}
	return asynq.NewTask(DeleteShare, payload)
}

// Handlers
func HandleDeleteShareTask(ctx context.Context, t *asynq.Task) error {
	id, err := t.Payload.GetString("share_id")
	if err != nil {
		return err
	}
	db, err := GetDatabase()
	if err != nil {
		return err
	}
	return db.Where("ID = ?", id).Delete(&Share{}).Error
}

func HandleShareEmailTask(ctx context.Context, t *asynq.Task) error {
	id, err := t.Payload.GetString("share_id")
	if err != nil {
		return err
	}
	return SendMail(id)
}
