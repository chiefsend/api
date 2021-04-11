package controllers

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chiefsend/api/background"
	m "github.com/chiefsend/api/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type HTTPError struct {
	Error   error
	Message string
	Code    int
}

type EndpointREST func(http.ResponseWriter, *http.Request) *HTTPError

func (fn EndpointREST) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *HTTPError, not os.Error.
		log.Println(fmt.Sprintf("%d: %s - %s", (*e).Code, (*e).Error.Error(), (*e).Message))
		if e.Code == 401 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please enter the password"`)
		}
		http.Error(w, fmt.Sprintf("%s - %s", (*e).Message, (*e).Error.Error()), (*e).Code)
	}
}

/////////////////////////////////
//////////// routes /////////////
/////////////////////////////////
func AllShares(w http.ResponseWriter, r *http.Request) *HTTPError {
	// get database
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}
	// see if (optional) admin key is provided to also return temporary shares
	admin, err := CheckBearerAuth(r)
	if err != nil {
		return &HTTPError{err, "can't check authorization header", 500}
	}
	// get shares
	var shares []m.Share
	if admin {
		err = db.Where("1=1").Find(&shares).Error
	} else {
		err = db.Where("is_public = 1 AND is_temporary = 0").Find(&shares).Error
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	// send shares
	return sendJSON(w, shares)
}

func GetShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}

	var share m.Share
	err = db.Preload("Attachments").Where("ID = ?", shareID).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	if share.IsTemporary == true {
		return &HTTPError{errors.New("share is not finalized"), "Share is not finalized", 403}
	}

	// auth
	if basic, err := CheckBasicAuth(r, share); err != nil || basic == false {
		return &HTTPError{err, "Unauthorized", 401}
	}

	return sendJSON(w, share)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) *HTTPError {
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}
	attID, err := uuid.Parse(vars["att"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	var att m.Attachment
	err = db.Where("id = ?", attID.String()).First(&att).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}

	if att.ShareID != shareID {
		return &HTTPError{errors.New("share doesent match attachment"), "share doesent match attachment", 404}
	}

	var share m.Share
	err = db.Where("id = ?", att.ShareID.String()).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	if share.IsTemporary == true {
		return &HTTPError{errors.New("share is not finalized"), "Share is not finalized", 403}
	}

	if ok, err := CheckBasicAuth(r, share); err != nil || ok == false {
		return &HTTPError{err, "Unauthorized", 401}
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", att.Filename))
	http.ServeFile(w, r, filepath.Join(os.Getenv("MEDIA_DIR"), "data", shareID.String(), attID.String()))
	return nil
}

func OpenShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}
	// parse body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &HTTPError{err, "Request does not contain a valid body", 400}
	}
	var newShare m.Share
	err = json.Unmarshal(reqBody, &newShare)
	if err != nil {
		return &HTTPError{err, "Can't parse body", 400}
	}
	// setup and store it
	newShare.Attachments = nil // dont want attachments yet
	newShare.IsTemporary = true
	err = db.Create(&newShare).Error
	if err != nil {
		return &HTTPError{err, "Can't create data", 500}
	}

	return sendJSON(w, newShare)
}

func CloseShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	var share m.Share
	err = db.Where("id = ?", shareID.String()).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	if share.IsTemporary == false { // already closed
		return nil
	}

	// move files to permanent location
	oldPath := filepath.Join(os.Getenv("MEDIA_DIR"), "temp", shareID.String())
	newPath := filepath.Join(os.Getenv("MEDIA_DIR"), "data", shareID.String())
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return &HTTPError{err, "Can't move directory", 500}
	}
	// set stuff permanent
	share.IsTemporary = false
	err = db.Save(&share).Error
	if err != nil {
		return &HTTPError{err, "Can't edit data", 500}
	}

	// send email
	mailTask := background.NewShareEmailTask(share)
	if err := background.EnqueueJob(mailTask, nil); err != nil {
		return &HTTPError{err, "Can't start send eMail background task", 500}
	}
	// delete share
	if share.Expires != nil {
		deleteTask := background.NewDeleteShareTask(share)
		if err := background.EnqueueJob(deleteTask, share.Expires); err != nil {
			return &HTTPError{err, "Can't start deleteShare task", 500}
		}
	}

	return sendJSON(w, share)
}

func UploadAttachment(w http.ResponseWriter, r *http.Request) *HTTPError {
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	var share m.Share
	err = db.Where("id = ?", shareID.String()).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	if share.IsTemporary == false { // TODO unless ADMIN key is passed
		return &HTTPError{errors.New("share is not finalized"), "Can't upload to finalized Shares.", 403}
	}

	// Parse file from body
	err = r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		return &HTTPError{err, "Request does not contain a valid body (parsing form)", 400}
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		return &HTTPError{err, "Request does not contain a valid body (parsing file)", 400}
	}
	defer file.Close()

	var att m.Attachment
	{
		// add database entry // TODO error handling for whole transaction
		db.Begin()
		att.ShareID = shareID
		att.Filename = handler.Filename
		att.Filesize = handler.Size
		db.Create(&att)

		// save file
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			return &HTTPError{err, "cant read file", 500}
		}
		err = ioutil.WriteFile(filepath.Join(os.Getenv("MEDIA_DIR"), "temp", shareID.String(), att.ID.String()), fileBytes, os.ModePerm)
		if err != nil {
			db.Rollback()
			return &HTTPError{err, "cant save file", 500}
		}
		db.Commit()
	}

	return sendJSON(w, att)
}

func DownloadZip(w http.ResponseWriter, r *http.Request) *HTTPError {
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	var share m.Share
	err = db.Preload("Attachments").Where("ID = ?", shareID).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	if share.IsTemporary == true {
		return &HTTPError{errors.New("share is not finalized"), "Share is not finalized", 403}
	}

	zipWriter := zip.NewWriter(w)
	for _, file := range share.Attachments {
		filePath := filepath.Join(os.Getenv("MEDIA_DIR"), "data", file.ShareID.String(), file.ID.String())

		fileToZip, err := os.Open(filePath)
		if err != nil {
			return &HTTPError{err, "error opening file", 500}
		}

		info, err := fileToZip.Stat()
		if err != nil {
			return &HTTPError{err, "error getting file info", 500}
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return &HTTPError{err, "error creating file header", 500}
		}

		header.Name = file.Filename
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return &HTTPError{err, "error creating header", 500}
		}
		if _, err := io.Copy(writer, fileToZip); err != nil {
			return &HTTPError{err, "error copying file into zip archive", 500}
		}
		if err := fileToZip.Close(); err != nil {
			return &HTTPError{err, "error closing zip archive", 500}
		}

	}
	err = zipWriter.Close()
	if err != nil {
		return &HTTPError{err, "error when closing zip", 500}
	}

	return nil
}

//////////////////////////////////////////
///////////// Admin Routes ///////////////
//////////////////////////////////////////
func DeleteShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}

	var share m.Share
	err = db.Preload("Attachments").Where("ID = ?", shareID).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}

	if auth, err := CheckBearerAuth(r); err != nil || auth == false {
		return &HTTPError{err, "Authentication Failed", 401}
	}

	// delete
	deleteTask := background.NewDeleteShareTask(share)
	if err := background.EnqueueJob(deleteTask, share.Expires); err != nil {
		return &HTTPError{err, "Can't start deleteShare task", 500}
	}

	return nil
}

func UpdateShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	// auth
	if ok, err := CheckBearerAuth(r); err != nil || ok == false {
		return &HTTPError{err, "Unauthorized", 401}
	}
	// parse url
	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}
	//  get database
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}
	// get share
	var share m.Share
	err = db.Where("ID = ?", shareID).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	// update share
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &HTTPError{err, "Request does not contain a valid body", 400}
	}
	err = json.Unmarshal(reqBody, &share)
	if err != nil {
		return &HTTPError{err, "Can't parse body", 400}
	}
	err = db.Save(&share).Error
	if err != nil {
		return &HTTPError{err, "Can't edit data", 500}
	}
	// finish
	return nil
}

func DeleteAttachment(w http.ResponseWriter, r *http.Request) *HTTPError {
	db, err := m.GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't connect to database", 500}
	}

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}
	attID, err := uuid.Parse(vars["att"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	var att m.Attachment
	err = db.Where("id = ?", attID.String()).First(&att).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}

	if att.ShareID != shareID {
		return &HTTPError{errors.New("share doesent match attachment"), "share doesent match attachment", 404}
	}

	var share m.Share
	err = db.Where("id = ?", att.ShareID.String()).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}

	// delete attachment
	err = db.Delete(&att).Error
	if err != nil {
		return &HTTPError{err, "can't delete attachment", 500}
	}

	return nil
}