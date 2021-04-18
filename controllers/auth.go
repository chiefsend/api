package controllers

import (
	"encoding/base64"
	"errors"
	"github.com/chiefsend/api/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strings"
)

// CheckBearerAuth returns true if the ADMIN_KEY is provided as Bearer token, false otherwise (or no token is included)
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
		return false, errors.New("invalid Authorization Header")
	}
	token, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return false, err
	}
	return string(token) == os.Getenv("ADMIN_KEY"), nil
}

// CheckBasicAuth returns true if data in basic auth header can unlock the share. Returns error if no Auth header is provided
func CheckBasicAuth(r *http.Request, share models.Share) (bool, error) {
	if share.Password.Valid {
		sid, pass, ok := r.BasicAuth()
		if !ok {
			return false, errors.New("invalid auth header")
		}
		if sid != share.ID.String() {
			return false, nil
		}

		if err := bcrypt.CompareHashAndPassword([]byte(share.Password.ValueOrZero()), []byte(pass)); err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
	return true, nil // no password
}
