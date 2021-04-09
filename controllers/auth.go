package controllers

import (
	"encoding/base64"
	"errors"
	"github.com/chiefsend/api/models"
	"net/http"
	"os"
	"strings"
)

func checkBearerAuth(r *http.Request) (bool, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false, errors.New("invalid Authorization Header")
	}
	const prefix = "Bearer "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return false, errors.New("invalid Authorization Header")
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return false, errors.New("invalid Authorization Header")
	}
	tok := string(c)
	return tok == os.Getenv("ADMIN_KEY"), nil
}

func checkBasicAuth(r *http.Request, share models.Share) (bool, error) {
	if share.Password != "" {
		sid, pass, _ := r.BasicAuth()
		if sid != share.ID.String() {
			return false, errors.New("wrong username")
		}
		if pass != share.Password {
			return false, errors.New("wrong password")
		}
	}
	return true, nil
}
