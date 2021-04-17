package controllers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/chiefsend/api/models"
	"golang.org/x/crypto/bcrypt"
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

// return true if data in basic auth header can unlock the share. Returns error if no Auth header is provided
func CheckBasicAuth(r *http.Request, share models.Share) (bool, error) {
	if share.Password != "" {
		sid, pass, ok := r.BasicAuth()
		fmt.Println(sid, pass, ok)
		if !ok {
			return false, errors.New("invalid auth header")
		}
		if sid != share.ID.String() {
			return false, nil
		}
		if !checkPasswordHash(pass, share.Password) {
			return false, nil
		}
	}
	return true, nil
}


// returns true if the password matches the hash
func checkPasswordHash(password, hash string) bool {
	fmt.Println(password, hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))  // FIXME doesen't work
	if err != nil {
		fmt.Println(err.Error())
	}
	return err == nil
}
