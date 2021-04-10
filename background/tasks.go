package background

import (
	"context"
	"github.com/chiefsend/api/mail"
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

func NewDeleteShareTask(share m.Share) *asynq.Task {
	payload := map[string]interface{}{"share_id": share.ID.String()}
	return asynq.NewTask(DeleteShare, payload)
}

// Handlers
func HandleDeleteShareTask(ctx context.Context, t *asynq.Task) error {
	db, err := m.GetDatabase()
	if err != nil {
		return err
	}

	id, err := t.Payload.GetString("share_id")
	if err != nil {
		return err
	}
	return db.Where("ID = ?", id).Delete(&m.Share{}).Error
}

func HandleShareEmailTask(ctx context.Context, t *asynq.Task) error {
	id, err := t.Payload.GetString("share_id")
	if err != nil {
		return err
	}
	return mail.SendMail(id)
}
