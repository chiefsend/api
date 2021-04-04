package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	g "github.com/chiefsend/api/globals"
	m "github.com/chiefsend/api/models"
	u "github.com/chiefsend/api/util"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllShares(t *testing.T) {
	u.Reset()
	r := mux.NewRouter()
	r.Handle("/shares", EndpointREST(AllShares)).Methods("GET")
	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		res, _ := http.Get(ts.URL + "/shares")
		body, _ := ioutil.ReadAll(res.Body)
		var erg []m.Share
		json.Unmarshal(body, &erg)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.EqualValues(t, u.Shares[1], erg[0])
	})
}

func TestGetShare(t *testing.T) {
	u.Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}", EndpointREST(GetShare)).Methods("GET")
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		header := map[string][]string{
			"Authorization": {"Basic NTcxM2QyMjgtYTA0Mi00NDZkLWE1ZTQtMTgzYjE5ZmE4MzJhOnRlc3QxMjM="}, // pw: test123
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", ts.URL, u.Shares[0].ID.String()), nil)
		req.Header = header
		res, _ := http.DefaultClient.Do(req)

		body, _ := ioutil.ReadAll(res.Body)
		var actual m.Share
		var expected m.Share

		json.Unmarshal(body, &actual)
		ex, _ := json.Marshal(u.Shares[0])
		json.Unmarshal(ex, &expected)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, expected, actual)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", ts.URL, u.Shares[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/4015a76b-09d0-402b-814f-bd9fa48ce8e1", ts.URL), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("not ready", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s", ts.URL, u.Shares[2].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})
}

func TestDownloadFile(t *testing.T) {
	u.Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}/attachment/{att}", EndpointREST(DownloadFile)).Methods("GET")
	defer ts.Close()

	t.Run("happy path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", ts.URL, u.Shares[0].ID.String(), u.Shares[0].Attachments[0].ID.String()), nil)
		req.Header = map[string][]string{
			"Authorization": {"Basic NTcxM2QyMjgtYTA0Mi00NDZkLWE1ZTQtMTgzYjE5ZmE4MzJhOnRlc3QxMjM="}, // pw: test123
		}
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		assert.FileExists(t, filepath.Join(g.Conf.MediaDir, "data", u.Shares[0].ID.String(), u.Shares[0].Attachments[0].ID.String()))
		expected, _ := ioutil.ReadFile(filepath.Join(g.Conf.MediaDir, "data", u.Shares[0].ID.String(), u.Shares[0].Attachments[0].ID.String()))
		assert.EqualValues(t, expected, body)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", ts.URL, u.Shares[0].ID.String(), u.Shares[0].Attachments[0].ID.String()), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/share/%s/attachment/%s", ts.URL, u.Shares[0].ID.String(), "0dd9a011-612b-4f33-99c0-bfd687021014"), nil)
		res, _ := http.DefaultClient.Do(req)
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestOpenShare(t *testing.T) {
	u.Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/shares", EndpointREST(OpenShare)).Methods("POST")
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
	u.Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}", EndpointREST(CloseShare)).Methods("POST")
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
	u.Reset()
	router := mux.NewRouter()
	ts := httptest.NewServer(router)
	router.Handle("/share/{id}/attachments", EndpointREST(UploadAttachment)).Methods("POST")
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
