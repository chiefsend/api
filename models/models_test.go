package models

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	db, err := GetDatabase()
	if err != nil {
		log.Fatal("database brok")
	}
	_ = db.AutoMigrate(&Share{})
	_ = db.AutoMigrate(&Attachment{})

	os.Exit(m.Run())
}

func TestCreateShare(t *testing.T) {
	uhr := time.Date(2020, 01, 01, 17, 17, 17, 324359102, time.UTC)
	var expected = Share{
		ID:            uuid.MustParse("1e21e633-7936-4dd5-9de5-43ed1c413d8a"),
		Name:          "Test1",
		Expires:       &uhr,
		DownloadLimit: 123,
		IsPublic:      true,
		Password:      "test123",
		Emails:        []string{"test@example.com"},
		IsTemporary:   true,
	}

	db, _ := GetDatabase()
	db.Create(&expected)

	var actual Share
	db.Where("id=?", "1e21e633-7936-4dd5-9de5-43ed1c413d8a").First(&actual)
	// assertions
	assert.Equal(t, expected, actual)
}

func TestReadShare(t *testing.T) {
	return
}

func TestUpdateShare(t *testing.T) {
	return
}

func TestDeleteShare(t *testing.T) {
	return
}

////////////////////////////////////////////
///////////// Attachment////////////////////
////////////////////////////////////////////
func TestCreateAttachment(t *testing.T) {
	return
}

func TestReadAttachment(t *testing.T) {
	return
}

func TestUpdateAttachment(t *testing.T) {
	return
}

func TestDeleteAttachment(t *testing.T) {
	return
}
