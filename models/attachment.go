package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

type Attachment struct {
	ID uuid.UUID `json:"id"  gorm:"primary_key"`

	Filename string `json:"filename"  gorm:"not null"`
	Filesize int64  `json:"filesize"  gorm:"not null; default:0"`

	ShareID uuid.UUID `json:"-"  gorm:"not null"`
}

func (att Attachment) String() string {
	indent, err := json.MarshalIndent(att, "", "    ")
	if err != nil {
		return "error printing attachment"
	}
	return string(indent)
}

func (att *Attachment) BeforeCreate(tx *gorm.DB) error {
	if att.ID.String() == "00000000-0000-0000-0000-000000000000" {
		uid, err := uuid.NewRandom()
		if err != nil {
			tx.Rollback()
			return err
		}
		att.ID = uid
	}
	return nil
}

func (att *Attachment) BeforeDelete(tx *gorm.DB) error {
	if err := os.Remove(filepath.Join(os.Getenv("MEDIA_DIR"), "data", att.ShareID.String(), att.ID.String())); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
