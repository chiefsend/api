package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chiefsend/api/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var shares = []models.Share{
	{
		ID:            uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
		Name:          "TestFinalPrivate",
		DownloadLimit: 100,
		IsPublic:      false,
		IsTemporary:   false,
		Password:      "test123",
		Emails:        []string{""},
		Attachments: []models.Attachment{
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

var url string
var db *gorm.DB

func TestMain(m *testing.M) {
	dab, err := models.GetDatabase()
	if err != nil {
		log.Fatal("database brok")
	}
	db = dab
	_ = db.AutoMigrate(&models.Share{})
	_ = db.AutoMigrate(&models.Attachment{})

	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Handle("/shares", EndpointREST(AllShares)).Methods("GET")
	router.Handle("/shares", EndpointREST(OpenShare)).Methods("POST")
	router.Handle("/share/{id}", EndpointREST(GetShare)).Methods("GET")
	router.Handle("/share/{id}", EndpointREST(CloseShare)).Methods("POST")
	router.Handle("/share/{id}/attachments", EndpointREST(UploadAttachment)).Methods("POST")
	router.Handle("/share/{id}/attachment/{att}", EndpointREST(DownloadFile)).Methods("GET")
	router.Handle("/share/{id}/zip", EndpointREST(DownloadZip)).Methods("GET")
	url = ts.URL

	code := m.Run()

	db.Where("1 = 1").Delete(&models.Share{})
	db.Where("1 = 1").Delete(&models.Attachment{})
	_ = os.RemoveAll(filepath.Join(os.Getenv("MEDIA_DIR"), "data"))
	_ = os.RemoveAll(filepath.Join(os.Getenv("MEDIA_DIR"), "temp"))

	os.Exit(code)
}

func TestAllShares(t *testing.T) {
	db.Create(&shares[1])
	db.Create(&shares[2])
	defer db.Delete(&shares[1])
	defer db.Delete(&shares[2])

	t.Run("happy path", func(t *testing.T) {
		res, _ := http.Get(url + "/shares")
		body, _ := ioutil.ReadAll(res.Body)
		var erg []models.Share
		_ = json.Unmarshal(body, &erg)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Len(t, erg, 1)
		assert.EqualValues(t, shares[1], erg[0])
	})
}

func TestGetShare(t *testing.T) {
	db.Create(&shares[0])
	defer db.Delete(&shares[0])
	db.Create(&shares[2])
	defer db.Delete(&shares[2])

	t.Run("happy path", func(t *testing.T) {
		header := map[string][]string{
			"Authorization": {"Basic NTcxM2QyMjgtYTA0Mi00NDZkLWE1ZTQtMTgzYjE5ZmE4MzJhOnRlc3QxMjM="}, // pw: test123
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, shares[0].ID.String()), nil)
		req.Header = header
		res, _ := http.DefaultClient.Do(req)

		body, _ := ioutil.ReadAll(res.Body)
		var actual models.Share
		var expected models.Share

		_ = json.Unmarshal(body, &actual)
		ex, _ := json.Marshal(shares[0])
		_ = json.Unmarshal(ex, &expected)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, expected, actual)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, shares[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/4015a76b-09d0-402b-814f-bd9fa48ce8e1", url), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("not ready", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, shares[2].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})
}

func TestDownloadFile(t *testing.T) {
	db.Create(&shares[0])
	defer db.Delete(&shares[0])
	if err := ioutil.WriteFile(filepath.Join(os.Getenv("MEDIA_DIR"), "data", shares[0].ID.String(), shares[0].Attachments[0].ID.String()), []byte("KEKW KEKW KEKW"), os.ModePerm); err == nil {
		defer os.Remove(filepath.Join(os.Getenv("MEDIA_DIR"), "data", shares[0].ID.String(), shares[0].Attachments[0].ID.String()))
	}

	t.Run("happy path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, shares[0].ID.String(), shares[0].Attachments[0].ID.String()), nil)
		req.Header = map[string][]string{
			"Authorization": {"Basic NTcxM2QyMjgtYTA0Mi00NDZkLWE1ZTQtMTgzYjE5ZmE4MzJhOnRlc3QxMjM="}, // pw: test123
		}
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		assert.FileExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "data", shares[0].ID.String(), shares[0].Attachments[0].ID.String()))
		expected, _ := ioutil.ReadFile(filepath.Join(os.Getenv("MEDIA_DIR"), "data", shares[0].ID.String(), shares[0].Attachments[0].ID.String()))
		assert.EqualValues(t, expected, body)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, shares[0].ID.String(), shares[0].Attachments[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, shares[0].ID.String(), "0dd9a011-612b-4f33-99c0-bfd687021014"), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestOpenShare(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var newShare = models.Share{
			ID:            uuid.MustParse("e5134044-2704-4864-85be-318fb158009f"),
			Name:          "TestOpenShare",
			Expires:       nil,
			DownloadLimit: 69,
			IsPublic:      false,
			Attachments: []models.Attachment{
				{
					ID:       uuid.MustParse("2b524827-9c3c-47e0-9277-8b51fd45b4bd"),
					Filename: "hallo.txt",
					Filesize: 123456,
					ShareID:  uuid.MustParse("e5134044-2704-4864-85be-318fb158009f"),
				},
			},
		}
		b, _ := json.Marshal(newShare)
		req, _ := http.NewRequest("POST", url+"/shares", strings.NewReader(string(b)))
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		var actual models.Share
		_ = json.Unmarshal(body, &actual)
		// checks
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.DirExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "temp", newShare.ID.String()))
		assert.NoDirExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "data", newShare.ID.String()))

		assert.Equal(t, newShare.ID, actual.ID)
		//assert.(t, newShare, actual)
	})

	t.Run("bad request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", url+"/shares", nil)
		res, _ := http.DefaultClient.Do(req)
		//body, _ := ioutil.ReadAll(res.Body)
		// checks
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestCloseShare(t *testing.T) {
	db.Create(&shares[2])
	defer db.Delete(&shares[2])

	t.Run("happy path", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/share/%s", url, shares[2].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("POST", url+"/share/e3334044-eeee-4864-85be-555fb158009f", nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestUploadAttachment(t *testing.T) {
	db.Create(&shares[2])
	defer db.Delete(&shares[2])

	t.Run("happy path", func(t *testing.T) {
		// build body (multipart/form-data)
		buff := bytes.Buffer{}
		writer := multipart.NewWriter(&buff)
		fw, _ := writer.CreateFormFile("file", "poggers.txt")
		_, _ = io.Copy(fw, strings.NewReader("POG POG POG\r\n"))
		// request
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/share/%s/attachments", url, shares[2].ID.String()), bytes.NewReader(buff.Bytes()))
		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, _ := http.DefaultClient.Do(req)

		b, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(b))

		var share models.Share
		db, _ := models.GetDatabase()
		db.Where("ID = ?", shares[2].ID.String()).First(&share)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Len(t, share.Attachments, 1)
	})

	t.Run("bad request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/share/%s/attachments", url, shares[2].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
