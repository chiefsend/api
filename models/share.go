package models

import (
	g "chiefsend-api/globals"
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
	Password      string     `json:"-"`
	Emails        []string   `json:"emails,omitempty" gorm:"-"`
	eMailsDB      string     `json:"-"`
	IsTemporary   bool       `json:"-"  gorm:"not null"`

	Attachments []Attachment `json:"files,omitempty"  gorm:"constraint:OnDelete:CASCADE"`
}


func (sh *Share) AfterFind(scope *gorm.DB) error {
	sh.Emails = strings.Split(sh.eMailsDB, ";")
	return nil
}

func (sh *Share) AfterCreate(scope *gorm.DB) error {
	if sh.ID.String() == "00000000-0000-0000-0000-000000000000" {
		uid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		sh.ID = uid
	}
	// create temporary dir
	if sh.IsTemporary == true {
		if err := os.MkdirAll(filepath.Join(g.Conf.MediaDir, "temp", sh.ID.String()), os.ModePerm); err != nil {
			return nil
		}
	} else {
		if err := os.MkdirAll(filepath.Join(g.Conf.MediaDir, "data", sh.ID.String()), os.ModePerm); err != nil {
			return nil
		}
	}
	//convert email adresses
	sh.eMailsDB = strings.Join(sh.Emails, ";")
	return nil
}

func (sh *Share) BeforeDelete(scope *gorm.DB) error {
	if sh.IsTemporary == false {
		return os.RemoveAll(filepath.Join(g.Conf.MediaDir, "data", sh.ID.String()))
	}
	return nil
}