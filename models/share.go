package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Share has many Attachments, ShareID is the foreign key
type Share struct {
	ID            uuid.UUID  `json:"id"  gorm:"primary_key"`
	Name          string     `json:"name,omitempty"`
	Expires       *time.Time `json:"expires,omitempty"`
	DownloadLimit int        `json:"download_limit,omitempty"`
	IsPublic      bool       `json:"is_public"  gorm:"not null; default:false; index"`
	Password      string     `json:"password,omitempty"`
	Emails        []string   `json:"emails,omitempty" gorm:"-"`
	EMailsDB      string     `json:"-"`
	IsTemporary   bool       `json:"is_temporary,omitempty" gorm:"not null"`

	Attachments []Attachment `json:"files,omitempty"  gorm:"constraint:OnDelete:CASCADE"`
}


func (sh *Share) AfterFind(tx *gorm.DB) error {
	sh.Emails = strings.Split(sh.EMailsDB, ";")
	return nil
}

func (sh *Share) BeforeCreate(tx *gorm.DB) error {
	// set uuid
	if sh.ID.String() == "00000000-0000-0000-0000-000000000000" {
		uid, err := uuid.NewRandom()
		if err != nil {
			tx.Rollback()
			return err
		}
		sh.ID = uid
	}
	// create temporary dir
	if sh.IsTemporary == true {
		if err := os.MkdirAll(filepath.Join(os.Getenv("MEDIA_DIR"), "temp", sh.ID.String()), os.ModePerm); err != nil {
			tx.Rollback()
			return nil
		}
	} else {
		if err := os.MkdirAll(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String()), os.ModePerm); err != nil {
			tx.Rollback()
			return nil
		}
	}
	//convert email adresses
	sh.EMailsDB = strings.Join(sh.Emails, ";")
	return nil
}

func (sh *Share) BeforeDelete(tx *gorm.DB) error {
	if sh.IsTemporary == false {
		if err := os.RemoveAll(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String())); err != nil {
			tx.Rollback()
			return err
		}
	}else {
		if err := os.RemoveAll(filepath.Join(os.Getenv("MEDIA_DIR"), "temp", sh.ID.String())); err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}