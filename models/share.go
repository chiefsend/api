package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"time"
)

// Share has many Attachments, ShareID is the foreign key
type Share struct {
	ID        uuid.UUID `json:"id"  gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`

	Name          null.String `json:"name,omitempty"`
	Expires       null.Time   `json:"expires,omitempty"`
	DownloadLimit null.Int    `json:"download_limit,omitempty"`
	IsPublic      bool        `json:"is_public"  gorm:"not null; default:false; index"`
	Password      null.String `json:"password,omitempty"`
	IsTemporary   bool        `json:"is_temporary,omitempty"`

	Attachments []Attachment `json:"files,omitempty"  gorm:"constraint:OnDelete:CASCADE"`
}

func (sh Share) String() string {
	indent, err := json.MarshalIndent(sh, "", "    ")
	if err != nil {
		return "error printing share"
	}
	return string(indent)
}

func (sh *Share) Secure() {
	if sh.Password.Valid {
		sh.Password.SetValid("")
	}
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
	// hash password
	if sh.Password.Valid {
		if hash, err := bcrypt.GenerateFromPassword([]byte(sh.Password.ValueOrZero()), bcrypt.DefaultCost); err != nil {
			return err
		} else {
			sh.Password.SetValid(string(hash))
		}
	}
	// return
	return nil
}

func (sh *Share) BeforeDelete(tx *gorm.DB) error {
	if sh.IsTemporary == false {
		if err := os.RemoveAll(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String())); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := os.RemoveAll(filepath.Join(os.Getenv("MEDIA_DIR"), "temp", sh.ID.String())); err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}
