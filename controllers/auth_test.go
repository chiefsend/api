package controllers

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
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

}
