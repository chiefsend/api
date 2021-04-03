package main

import (
	"context"
	g "github.com/chiefsend/api/globals"
	m "github.com/chiefsend/api/models"
	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	DeleteShare = "share:delete"
	ShareEmail  = "email:share"
)

// Tasks
func NewShareEmailTask(share m.Share) *asynq.Task {
	payload := map[string]interface{}{"share_id": share.ID.String()}
	return asynq.NewTask(ShareEmail, payload)
}

func NewDeleteShareTaks(share m.Share) *asynq.Task {
	payload := map[string]interface{}{"share_id": share.ID.String()}
	return asynq.NewTask(DeleteShare, payload)
}

// Handlers
func HandleDeleteShareTask(ctx context.Context, t *asynq.Task) error {
	id, err := t.Payload.GetString("share_id")
	if err != nil {
		return err
	}
	return g.Db.Where("ID = ?", id).Delete(&m.Share{}).Error
}

func HandleShareEmailTask(ctx context.Context, t *asynq.Task) error {
	id, err := t.Payload.GetString("share_id")
	if err != nil {
		return err
	}
	return SendMail(id)
}
