package controllers

import (
	"encoding/base64"
	m "github.com/chiefsend/api/models"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"os"
	"testing"
)

func TestCheckBearerAuth(t *testing.T) {
	t.Run("with header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/random", nil)
		req.Header.Set("Authorization", "Bearer " + base64.StdEncoding.EncodeToString([]byte(os.Getenv("ADMIN_KEY"))))
		ok, err := CheckBearerAuth(req)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("without header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/random", nil)
		ok, err := CheckBearerAuth(req)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

}

func TestCheckBasicAuth(t *testing.T) {
	sh := m.Share{
		Password:    null.StringFrom("test123"),
	}
	db.Create(&sh)
	defer db.Delete(&sh)

	t.Run("happy path", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("GET", "/random", nil)
		req.SetBasicAuth(sh.ID.String(), "test123")
		ok, err := CheckBasicAuth(req, sh)
		// assertions
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("without header", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("GET", "/random", nil)
		ok, err := CheckBasicAuth(req, sh)
		// assertions
		assert.NotNil(t, err)
		assert.False(t, ok)
	})

	t.Run("wrong username/shareID", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("GET", "/random", nil)
		req.SetBasicAuth("trash", sh.Password.ValueOrZero())
		ok, err := CheckBasicAuth(req, sh)
		// assertions
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("wrong password", func(t *testing.T) {
		// request
		req, _ := http.NewRequest("GET", "/random", nil)
		req.SetBasicAuth(sh.ID.String(), "trash")
		ok, err := CheckBasicAuth(req, sh)
		// assertions
		assert.NotNil(t, err)
		assert.False(t, ok)
	})
}
