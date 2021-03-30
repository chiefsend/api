package main

import (
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
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
	EMailsDB      string     `json:"-"`
	IsTemporary   bool       `json:"-"  gorm:"not null"`

	Attachments []Attachment `json:"files,omitempty"  gorm:"constraint:OnDelete:CASCADE"`
}

type Attachment struct {
	ID          uuid.UUID `json:"id"  gorm:"primary_key"`
	Filename    string    `json:"filename"  gorm:"not null"`
	Filesize    int64     `json:"filesize"  gorm:"not null; default:0"`
	IsEncrypted bool      `json:"-"  gorm:"not null; default:false"`

	ShareID uuid.UUID `json:"-"  gorm:"not null"`
}

/////////////////////////////////
/////////// Hoooks //////////////
/////////////////////////////////
func (sh *Share) AfterFind(scope *gorm.DB) error {
	sh.Emails = strings.Split(sh.EMailsDB, ";")
	return nil
}

func (sh *Share) BeforeCreate(scope *gorm.DB) error {
	if sh.ID.String() == "00000000-0000-0000-0000-000000000000" {
		uid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		sh.ID = uid
	}
	// create temporary dir
	if sh.IsTemporary == true {
		if err := os.MkdirAll(filepath.Join(config.mediaDir, "temp", sh.ID.String()), os.ModePerm); err != nil {
			return nil
		}

	} else {
		if err := os.MkdirAll(filepath.Join(config.mediaDir, "data", sh.ID.String()), os.ModePerm); err != nil {
			return nil
		}
	}
	//convert email adresses
	sh.EMailsDB = strings.Join(sh.Emails, ";")
	return nil
}

func (sh *Share) BeforeDelete(scope *gorm.DB) error {
	if sh.IsTemporary == false {
		return os.RemoveAll(filepath.Join(config.mediaDir, "data", sh.ID.String()))
	}
	return nil
}

func (att *Attachment) BeforeCreate(scope *gorm.DB) error {
	if att.ID.String() == "00000000-0000-0000-0000-000000000000" {
		uid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		att.ID = uid
	}
	return nil
}

func (att *Attachment) BeforeDelete(scope *gorm.DB) error {
	return os.Remove(filepath.Join(config.mediaDir, "data", att.ShareID.String(), att.ID.String()))
}

/////////////////////////////////
/////////// Functions ///////////
/////////////////////////////////
func GetDatabase() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URI")
	split := strings.SplitN(dsn, ":", 2)
	if split[0] == "postgres" {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else if split[0] == "sqlserver" {
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	} else if split[0] == "sqlite" || split[0] == "file" {
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}
	return nil, nil
}
