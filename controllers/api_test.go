package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chiefsend/api/background"
	m "github.com/chiefsend/api/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
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

// Parse and Reparse Share in JSON to simulate transmission
func parseShare(sh m.Share) m.Share {
	sh.Secure()
	body, err := json.Marshal(sh)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
	var e m.Share
	if err := json.Unmarshal(body, &e); err != nil {
		log.Fatal(err)
	}
	return e
}

var url string
var db *gorm.DB

func TestMain(mt *testing.M) {
	_ = os.Setenv("ADMIN_KEY", "testkey123")

	dab, err := m.GetDatabase()
	if err != nil {
		log.Fatal("database brok")
	}
	db = dab
	_ = db.AutoMigrate(&m.Share{})
	_ = db.AutoMigrate(&m.Attachment{})

	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	configureRoutes(router)
	url = ts.URL

	os.Exit(mt.Run())
}

////////////////////////////////////////
////////////// TEST CASES //////////////
////////////////////////////////////////
func TestAllShares(t *testing.T) {
	shares := []m.Share{
		{
			ID:       uuid.MustParse("f43b0e48-13cc-4c6c-8a23-3a18a670effd"),
			IsPublic: true,
		},
		{
			ID:       uuid.MustParse("a558aca3-fb40-400b-8dc6-ae49c705c791"),
			IsPublic: false,
		},
	}
	db.Create(&shares[0])
	db.Create(&shares[1])
	defer db.Delete(&shares[0])
	defer db.Delete(&shares[1])

	t.Run("happy path", func(t *testing.T) {
		// request
		res, _ := http.Get(url + "/shares")
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual []m.Share
		var expected = []m.Share{parseShare(shares[0])}
		_ = json.Unmarshal(body, &actual)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Len(t, actual, len(expected))
		assert.Equal(t, expected, actual)
	})

	t.Run("with admin key", func(t *testing.T) {
		// do request
		req, _ := http.NewRequest("GET", fmt.Sprint(url, "/shares"), nil)
		req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual []m.Share
		var expected = []m.Share{parseShare(shares[0]), parseShare(shares[1])}
		_ = json.Unmarshal(body, &actual)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Len(t, actual, len(expected))
		assert.Equal(t, expected, actual)
	})
}

func TestGetShare(t *testing.T) {
	shares := []m.Share{
		{
			ID:          uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
			IsTemporary: false,
			Password:    null.StringFrom("test123"),
		},
		{
			ID:          uuid.MustParse("a558aca3-fb40-400b-8dc6-ae49c705c791"),
			IsTemporary: true,
		},
	}
	db.Create(&shares[0])
	db.Create(&shares[1])
	defer db.Delete(&shares[0])
	defer db.Delete(&shares[1])

	t.Run("happy path", func(t *testing.T) {
		// request
		sh := shares[0]
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), nil)
		req.SetBasicAuth(sh.ID.String(), "test123")
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual m.Share
		_ = json.Unmarshal(body, &actual)
		var expected = parseShare(shares[0])
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
		res, _ := http.Get(fmt.Sprintf("%s/share/%s", url, shares[1].ID.String()))
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})

	t.Run("with admin key", func(t *testing.T) {
		// request
		sh := shares[0]
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual m.Share
		_ = json.Unmarshal(body, &actual)
		var expected = parseShare(shares[0])
		// assert
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, expected, actual)
	})
}

func TestDownloadFile(t *testing.T) {
	sh := m.Share{
		ID:            uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
		IsPublic:      false,
		IsTemporary:   false,
		Password:    null.StringFrom("test123"),
		Attachments: []m.Attachment{
			{
				ID:       uuid.MustParse("913134c0-894f-4c4d-b545-92ec373168b1"),
				Filename: "kekw.txt",
				Filesize: 123456,
				ShareID:  uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
			},
		},
	}
	db.Create(&sh)
	defer db.Delete(&sh)
	if err := ioutil.WriteFile(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()), []byte("KEKW KEKW KEKW"), os.ModePerm); err == nil {
		defer os.Remove(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
	}

	t.Run("happy path", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, sh.ID.String(), sh.Attachments[0].ID.String()), nil)
		req.SetBasicAuth(sh.ID.String(), "test123")
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		expected, _ := ioutil.ReadFile(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
		// assertions
		assert.FileExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
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

	t.Run("with admin key", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", url, sh.ID.String(), sh.Attachments[0].ID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		expected, _ := ioutil.ReadFile(filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
		// assert
		assert.FileExists(t, filepath.Join(os.Getenv("MEDIA_DIR"), "data", sh.ID.String(), sh.Attachments[0].ID.String()))
		assert.EqualValues(t, expected, body)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})
}

func TestOpenShare(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// request
		var newShare = m.Share{
			ID:            uuid.MustParse("e5134044-2704-4864-85be-318fb158009f"),
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
		req, _ := http.NewRequest("POST", url+"/shares", strings.NewReader(string(b)))
		res, _ := http.DefaultClient.Do(req)
		// parse
		body, _ := ioutil.ReadAll(res.Body)
		var actual m.Share
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
	sh := m.Share{
		ID:          uuid.MustParse("a558aca3-fb40-400b-8dc6-ae49c705c791"),
		IsTemporary: true,
	}
	db.Create(&sh)
	defer db.Delete(&sh)

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
	sh := m.Share{
		ID:          uuid.MustParse("a558aca3-fb40-400b-8dc6-ae49c705c791"),
		IsTemporary: true,
	}
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
		var actual m.Attachment
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
	sh := m.Share{
		ID: uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
	}
	db.Create(&sh)
	defer db.Delete(&sh)
	go background.StartBackgroundWorkers()
	defer background.StopBackgroundWorkers()

	t.Run("happy path", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("unauthorized", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		// assertions
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})
}

func TestUpdateShare(t *testing.T) {
	var sh = m.Share{
		ID: uuid.MustParse("f43b0e48-13cc-4c6c-8a23-3a18a670effd"),
	}
	db.Create(&sh)
	defer db.Delete(&sh)

	t.Run("happy path", func(t *testing.T) {
		// request
		sh.Name.SetValid("UpdatedName")
		b, _ := json.Marshal(sh)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/share/%s", url, sh.ID.String()), bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// parse
		var actual m.Share
		db.Where("ID = ?", sh.ID.String()).First(&actual)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, sh, actual)
	})
}

func TestDeleteAttachment(t *testing.T) {
	sh := m.Share{
		ID:            uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
		IsPublic:      false,
		IsTemporary:   false,
		Password:      null.StringFrom("test123"),
		Attachments: []m.Attachment{
			{
				ID:       uuid.MustParse("913134c0-894f-4c4d-b545-92ec373168b1"),
				Filename: "kekw.txt",
				Filesize: 123456,
				ShareID:  uuid.MustParse("5713d228-a042-446d-a5e4-183b19fa832a"),
			},
		},
	}
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
		req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		res, _ := http.DefaultClient.Do(req)
		// assertions
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.NoFileExists(t, path)
	})
}
