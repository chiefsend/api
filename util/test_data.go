package util

import (
	g "github.com/chiefsend/api/globals"
	m "github.com/chiefsend/api/models"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var Shares = []m.Share{
	{
		ID:            uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
		Name:          "TestFinalPrivate",
		DownloadLimit: 100,
		IsPublic:      false,
		IsTemporary:   false,
		Password:      "test123",
		Emails:        []string{""},

		Attachments: []m.Attachment{
			{
				ID:       uuid.MustParse("913134c0-894f-4c4d-b545-92ec373168b1"),
				Filename: "kekw.txt",
				Filesize: 123456,
				ShareID:  uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
			},
		},
	},
	{
		ID:          uuid.MustParse("f43b0e48-13cc-4c6c-8a23-3a18a670effd"),
		Name:        "TestFinalPublic",
		IsPublic:    true,
		IsTemporary: false,
		Emails:      []string{""},
	},
	{
		ID:            uuid.MustParse("a558aca3-fb40-400b-8dc6-ae49c705c791"),
		Name:          "TestTemporary",
		DownloadLimit: 300,
		IsPublic:      true,
		IsTemporary:   true,
		Emails:        []string{""},
	},
}

func Reset() {
	//g.LoadConfig()
	db, err := m.GetDatabase()
	if err != nil {
		log.Fatal("database brok")
	}
	db.AutoMigrate(&m.Share{})
	db.AutoMigrate(&m.Attachment{})
	// delete everything
	db.Where("1 = 1").Delete(&m.Share{})
	db.Where("1 = 1").Delete(&m.Attachment{})
	os.RemoveAll(filepath.Join(g.Conf.MediaDir, "data"))
	os.RemoveAll(filepath.Join(g.Conf.MediaDir, "temp"))
	// create everything
	for _, sh := range Shares {
		db.Create(&sh)
	}
	os.MkdirAll(filepath.Join(g.Conf.MediaDir, "data"), os.ModePerm)
	os.MkdirAll(filepath.Join(g.Conf.MediaDir, "temp"), os.ModePerm)
	// testfiles
	ioutil.WriteFile(filepath.Join(g.Conf.MediaDir, "data", Shares[0].ID.String(), Shares[0].Attachments[0].ID.String()), []byte("KEKW KEKW KEKW"), os.ModePerm)
}
