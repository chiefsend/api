package main

import (
	"bytes"
	g "chiefsend-api/globals"
	m "chiefsend-api/models"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
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

// DATA
var shares = []m.Share{
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
	g.LoadConfig()
	database, err := m.GetDatabase()
	if err != nil {
		log.Fatal("database brok")
	}
	g.Db = database
	g.Db.AutoMigrate(&m.Share{})
	g.Db.AutoMigrate(&m.Attachment{})
	// delete everything
	g.Db.Where("1 = 1").Delete(&m.Share{})
	g.Db.Where("1 = 1").Delete(&m.Attachment{})
	os.RemoveAll(filepath.Join(g.Conf.MediaDir, "data"))
	os.RemoveAll(filepath.Join(g.Conf.MediaDir, "temp"))
	// create everything
	for _, sh := range shares {
		g.Db.Create(&sh)
	}
	os.MkdirAll(filepath.Join(g.Conf.MediaDir, "data"), os.ModePerm)
	os.MkdirAll(filepath.Join(g.Conf.MediaDir, "temp"), os.ModePerm)
	// testfiles
	ioutil.WriteFile(filepath.Join(g.Conf.MediaDir, "data", shares[0].ID.String(), shares[0].Attachments[0].ID.String()), []byte("KEKW KEKW KEKW"), os.ModePerm)
}

/////////////////////////////////////
/////////////// TEST ////////////////
/////////////////////////////////////
func TestAllShares(t *testing.T) {
	Reset()
	r := mux.NewRouter()
	r.Handle("/shares", endpointREST(AllShares)).Methods("GET")
	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		res, _ := http.Get(ts.URL + "/shares")
		body, _ := ioutil.ReadAll(res.Body)
		var erg []m.Share
		json.Unmarshal(body, &erg)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.EqualValues(t, shares[1], erg[0])
	})
}

func TestGetShare(t *testing.T) {
	Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}", endpointREST(GetShare)).Methods("GET")
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		header := map[string][]string{
			"Authorization": {"Basic NTcxM2QyMjgtYTA0Mi00NDZkLWE1ZTQtMTgzYjE5ZmE4MzJhOnRlc3QxMjM="}, // pw: test123
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", ts.URL, shares[0].ID.String()), nil)
		req.Header = header
		res, _ := http.DefaultClient.Do(req)

		body, _ := ioutil.ReadAll(res.Body)
		var actual m.Share
		var expected m.Share

		json.Unmarshal(body, &actual)
		ex, _ := json.Marshal(shares[0])
		json.Unmarshal(ex, &expected)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, expected, actual)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", ts.URL, shares[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/4015a76b-09d0-402b-814f-bd9fa48ce8e1", ts.URL), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("not ready", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", ts.URL, shares[2].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})
}

func TestDownloadFile(t *testing.T) {
	Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}/attachment/{att}", endpointREST(DownloadFile)).Methods("GET")
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", ts.URL, shares[0].ID.String(), shares[0].Attachments[0].ID.String()), nil)
		req.Header = map[string][]string{
			"Authorization": {"Basic NTcxM2QyMjgtYTA0Mi00NDZkLWE1ZTQtMTgzYjE5ZmE4MzJhOnRlc3QxMjM="}, // pw: test123
		}
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		assert.FileExists(t, filepath.Join(g.Conf.MediaDir, "data", shares[0].ID.String(), shares[0].Attachments[0].ID.String()))
		expected, _ := ioutil.ReadFile(filepath.Join(g.Conf.MediaDir, "data", shares[0].ID.String(), shares[0].Attachments[0].ID.String()))
		assert.EqualValues(t, expected, body)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", ts.URL, shares[0].ID.String(), shares[0].Attachments[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", ts.URL, shares[0].ID.String(), "0dd9a011-612b-4f33-99c0-bfd687021014"), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestOpenShare(t *testing.T) {
	Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/shares", endpointREST(OpenShare)).Methods("POST")
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		var newShare = m.Share{
			ID:            uuid.MustParse("e5134044-2704-4864-85be-318fb158009f"),
			Name:          "TestOpenShare",
			Expires:       nil,
			DownloadLimit: 69,
			IsPublic:      false,
			Attachments: []m.Attachment{
				{
					ID:       uuid.MustParse("2b524827-9c3c-47e0-9277-8b51fd45b4bd"),
					Filename: "hallo.txt",
					Filesize: 123456,
					ShareID:  uuid.MustParse("e5134044-2704-4864-85be-318fb158009f"),
				},
			},
		}
		b, _ := json.Marshal(newShare)
		req, _ := http.NewRequest("POST", ts.URL+"/shares", strings.NewReader(string(b)))
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		var actual m.Share
		json.Unmarshal(body, &actual)
		// checks
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.DirExists(t, filepath.Join(g.Conf.MediaDir, "temp", newShare.ID.String()))
		assert.NoDirExists(t, filepath.Join(g.Conf.MediaDir, "data", newShare.ID.String()))

		assert.Equal(t, newShare.ID, actual.ID)
		//assert.(t, newShare, actual)
	})

	t.Run("bad request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/shares", nil)
		res, _ := http.DefaultClient.Do(req)
		//body, _ := ioutil.ReadAll(res.Body)
		// checks
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestCloseShare(t *testing.T) {
	Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}", endpointREST(CloseShare)).Methods("POST")
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/share/a558aca3-fb40-400b-8dc6-ae49c705c791", nil)
		res, _ := http.DefaultClient.Do(req)

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/share/e3334044-eeee-4864-85be-555fb158009f", nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("bad request", func(t *testing.T) {

	})
}

func TestUploadAttachment(t *testing.T) {
	Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}/attachments", endpointREST(UploadAttachment)).Methods("POST")
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		fw, _ := writer.CreateFormFile("file", "poggers.txt")
		io.Copy(fw, strings.NewReader("POG POG POG POG"))
		writer.Close()

		req, _ := http.NewRequest("POST", ts.URL+"/share/a558aca3-fb40-400b-8dc6-ae49c705c791/attachments", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, _ := http.DefaultClient.Do(req)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {

	})

	t.Run("bad request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/share/a558aca3-fb40-400b-8dc6-ae49c705c791/attachments", nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("forbidden", func(t *testing.T) {

	})
}
