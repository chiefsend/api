package controllers

import (
	"encoding/base64"
	"errors"
	"github.com/chiefsend/api/models"
	"net/http"
	"os"
	"strings"
)

func CheckBearerAuth(r *http.Request) (bool, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false, nil
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return false, nil
	}
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return false, errors.New("Invalid Authorization Header")
	}
	token, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return false, err
	}
	return string(token) == os.Getenv("ADMIN_KEY"), nil
}

func CheckBasicAuth(r *http.Request, share models.Share) (bool, error) {
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
