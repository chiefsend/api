package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chiefsend/api/background"
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
	_ = os.Setenv("ADMIN_KEY", "testkey123")

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
	router.Handle("/share/{id}", EndpointREST(DeleteShare)).Methods("DELETE")
	router.Handle("/share/{id}", EndpointREST(UpdateShare)).Methods("PUT")
	router.Handle("/share/{id}/attachments", EndpointREST(UploadAttachment)).Methods("POST")
	router.Handle("/share/{id}/attachment/{att}", EndpointREST(DownloadFile)).Methods("GET")
	router.Handle("/share/{id}/attachment/{att}", EndpointREST(DeleteAttachment)).Methods("DELETE")
	router.Handle("/share/{id}/zip", EndpointREST(DownloadZip)).Methods("GET")
	url = ts.URL

	os.Exit(m.Run())
}

////////////////////////////////////////
////////////// TEST CASES //////////////
////////////////////////////////////////
func TestAllShares(t *testing.T) {
	db.Create(&shares[1])
	db.Create(&shares[2])
	defer db.Delete(&shares[1])
	defer db.Delete(&shares[2])

	t.Run("happy path", func(t *testing.T) {
		// request
		res, _ := http.Get(url + "/shares")
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual []models.Share
		var expected = []models.Share{shares[1]}
		_ = json.Unmarshal(body, &actual)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Len(t, actual, 1)
		assert.Equal(t, expected, actual)
	})
}

func TestGetShare(t *testing.T) {
	db.Create(&shares[0])
	defer db.Delete(&shares[0])
	db.Create(&shares[2])
	defer db.Delete(&shares[2])

	t.Run("happy path", func(t *testing.T) {
		// reqeust
		sh := shares[0]
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), nil)
		req.SetBasicAuth(sh.ID.String(), sh.Password)
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual models.Share
		var expected models.Share
		_ = json.Unmarshal(body, &actual)
		ex, _ := json.Marshal(shares[0])
		_ = json.Unmarshal(ex, &expected)
		// assert
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, expected, actual)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, shares[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, uuid.New().String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("not ready", func(t *testing.T) {
		res, _ := http.Get( fmt.Sprintf("%s/share/%s", url, shares[2].ID.String()))
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})
}

func TestDownloadFile(t *testing.T) {
	sh := shares[0]
	db.Create(&sh)
	defer db.Delete(&sh)
	if err := ioutil.WriteFile(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()), []byte("KEKW KEKW KEKW"), os.ModePerm); err == nil {
		defer os.Remove(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
	}

	t.Run("happy path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, sh.ID.String(), sh.Attachments[0].ID.String()), nil)
		req.SetBasicAuth(sh.ID.String(), sh.Password)

		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		assert.FileExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
		expected, _ := ioutil.ReadFile(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
		assert.EqualValues(t, expected, body)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, sh.ID.String(), sh.Attachments[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, sh.ID.String(), uuid.New().String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestOpenShare(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// request
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
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual models.Share
		_ = json.Unmarshal(body, &actual)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.DirExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "temp", newShare.ID.String()))
		assert.NoDirExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "data", newShare.ID.String()))
		assert.Equal(t, newShare.ID, actual.ID)
	})

	t.Run("bad request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", url+"/shares", nil)
		res, _ := http.DefaultClient.Do(req)
		// assert
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestCloseShare(t *testing.T) {
	sh := shares[2]
	db.Create(&sh)

	t.Run("happy path", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/share/%s", url, uuid.New().String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Cleanup(func() {
		db.Delete(&sh)
		_ = os.RemoveAll(os.Getenv("MEDIA_DIR"))
	})
}

func TestUploadAttachment(t *testing.T) {
	sh := shares[2]
	db.Create(&sh)
	defer db.Delete(&sh)

	t.Run("happy path", func(t *testing.T) {
		// build body (multipart/form-data)
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)
		fw, _ := writer.CreateFormFile("file", "poggers.txt")
		_, _ = io.Copy(fw, strings.NewReader("POG POG POG"))
		writer.Close()
		// request
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/share/%s/attachments", url, sh.ID.String()), bytes.NewReader(b.Bytes()))
		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual models.Attachment
		_ = json.Unmarshal(body, &actual)
		// assertions
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.EqualValues(t, "poggers.txt", actual.Filename)
	})

	t.Run("bad request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/share/%s/attachments", url, sh.ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestDeleteShare(t *testing.T) {
	sh := shares[0]
	db.Create(&sh)
	defer db.Delete(&sh)
	go background.StartBackgroundWorkers()
	defer background.StopBackgroundWorkers()
	
	t.Run("happy path", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), nil)
		req.Header.Set("Authorization", "Bearer " + base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func TestUpdateShare(t *testing.T) {
	var sh = shares[1]
	db.Create(&sh)
	defer db.Delete(&sh)

	t.Run("happy path", func(t *testing.T) {
		// request
		sh.Name = "UpdatedName"
		b, _ := json.Marshal(sh)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer " + base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// parse
		var actual models.Share
		db.Where("ID = ?", sh.ID.String()).First(&actual)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, sh, actual)
	})
}

func TestDeleteAttachment(t *testing.T) {
	sh := shares[0]
	db.Create(&sh)
	defer db.Delete(&sh)
	path := filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String())
	if err := ioutil.WriteFile(path, []byte("KEKW KEKW KEKW"), os.ModePerm); err == nil {
		defer os.Remove(path)
	}

	t.Run("happy path", func(t *testing.T) {
		assert.FileExists(t, path)
		// request
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/share/%s/attachment/%s", url, sh.ID.String(), sh.Attachments[0].ID.String()), nil)
		req.Header.Set("Authorization", "Bearer " + base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.NoFileExists(t, path)
	})
}