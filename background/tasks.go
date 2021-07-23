package background

import (
	"context"
	m "github.com/chiefsend/api/models"
	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	DeleteShare      = "share:delete"
	ContinuousDelete = "continuous:delete"
)

// Tasks
func NewDeleteShareTask(share m.Share) *asynq.Task {
	payload := map[string]interface{}{"share_id": share.ID.String()}
	return asynq.NewTask(DeleteShare, payload)
}

func NewContinuousDeleteTask() *asynq.Task {
	return asynq.NewTask(ContinuousDelete, map[string]interface{}{})
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

func HandleContinuousDeleteTask(ctx context.Context, t *asynq.Task) error {
	db, err := m.GetDatabase()
	if err != nil {
		return err
	}

	//var tm = time.Now()
	var shares []m.Share
	if err := db.Find(&shares).Error; err != nil {
		return err
	}

	for _, sh := range shares {
		if sh.IsTemporary { // && tm.After(sh.CreatedAt.Add(24*time.Hour)) { // only delete open jobs if older than one day FIXME
			if err := db.Delete(&sh).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
