package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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

type endpointREST func(http.ResponseWriter, *http.Request) *HTTPError

func (fn endpointREST) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *HTTPError, not os.Error.
		if e.Code == 401 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please enter the password"`)
		}
		http.Error(w, fmt.Sprintf("%s - %s", (*e).Message, (*e).Error.Error()), (*e).Code)
	}
}

/////////////////////////////////
//////////// routes /////////////
/////////////////////////////////
func AllShares(w http.ResponseWriter, _ *http.Request) *HTTPError {
	fmt.Println("AllShares")
	db, err := GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't get database", 500}
	}
	var shares []Share
	err = db.Where("is_public = ? AND is_temporary = 0", 1).Find(&shares).Error
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	return SendJSON(w, shares)
}

func GetShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	fmt.Println("Get Share")
	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	db, err := GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't get database", 500}
	}

	var share Share
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

	// auth
	if share.Password != "" {
		sid, pass, _ := r.BasicAuth()
		if sid != share.ID.String() {
			return &HTTPError{errors.New("unauthorized"), "wrong username", 401}
		}
		if pass != share.Password {
			return &HTTPError{errors.New("unauthorized"), "wrong password", 401}
		}
	}

	return SendJSON(w, share)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) *HTTPError {
	fmt.Println("Download file")

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}
	attID, err := uuid.Parse(vars["att"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	db, err := GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't get database", 500}
	}

	var att Attachment
	err = db.Where("id = ?", attID.String()).First(&att).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}

	var share Share
	err = db.Where("id = ?", att.ShareID.String()).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}

	// auth
	if share.Password != "" {
		sid, pass, _ := r.BasicAuth()
		if sid != share.ID.String() {
			return &HTTPError{errors.New("unauthorized"), "wrong username", 401}
		}
		if pass != share.Password {
			return &HTTPError{errors.New("unauthorized"), "wrong password", 401}
		}
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", att.Filename))
	http.ServeFile(w, r, filepath.Join(config.mediaDir, "data", shareID.String(), attID.String()))
	return nil
}

func OpenShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	fmt.Println("OpenShare")

	db, err := GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't get database", 500}
	}

	// parse body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &HTTPError{err, "Request does not contain a valid body", 400}
	}
	var newShare Share
	err = json.Unmarshal(reqBody, &newShare)
	if err != nil {
		return &HTTPError{err, "Can't parse body", 400}
	}
	newShare.Attachments = nil // dont want attachments yet

	// create temporary db entrie
	newShare.IsTemporary = true
	err = db.Create(&newShare).Error
	if err != nil {
		return &HTTPError{err, "Can't create data", 500}
	}

	return SendJSON(w, newShare)
}

func CloseShare(w http.ResponseWriter, r *http.Request) *HTTPError {
	fmt.Println("CloseShare")

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	db, err := GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't get database", 500}
	}

	// get stuff
	var share Share
	err = db.Where("id = ?", shareID.String()).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}

	// move files to permanent location
	oldPath := filepath.Join(config.mediaDir, "temp", shareID.String())
	newPath := filepath.Join(config.mediaDir, "data", shareID.String())
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

	// TODO send mail
	SendMail(share)
	// TODO background job
	//{
	//	job, err := enqueuer.Enqueue("DeleteShare", nil)
	//	////deleteIn := time.Now().Sub(*share.Expires)
	//	//deleteIn := 1
	//	//job, err := enqueuer.EnqueueIn("DeleteShare", int64(deleteIn), map[string]interface{}{
	//	//	"ShareID": share.ID,
	//	//})
	//	if err != nil {
	//		return &HTTPError{err, "Error creating background job", 500}
	//	}
	//	PrettyPrint(job)
	//}

	return SendJSON(w, share)
}

func UploadAttachment(w http.ResponseWriter, r *http.Request) *HTTPError {
	fmt.Println("UploadTest")

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	db, err := GetDatabase()
	if err != nil {
		return &HTTPError{err, "Can't get database", 500}
	}

	// get share
	var share Share
	err = db.Where("id = ?", shareID.String()).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return &HTTPError{err, "Can't fetch data", 500}
	}
	if share.IsTemporary != true {
		return &HTTPError{errors.New("share is not finalized"), "Can't upload to finalized Shares.", 403}
	}

	// Parse file from body
	err = r.ParseMultipartForm(config.chunkSize) // Maximum 10 MB in RAM
	if err != nil {
		return &HTTPError{err, "Request does not contain a valid body (parsing form)", 400}
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		return &HTTPError{err, "Request does not contain a valid body (parsing file)", 400}
	}
	defer file.Close()

	var att Attachment
	{
		// add db entry TODO fehlerbehandlung für die ganze transaction
		db.Begin()
		sid, err := uuid.Parse(shareID.String())
		if err != nil {
			return &HTTPError{err, "foreign key shareID not valid", 500}
		}
		att.ShareID = sid
		att.Filename = handler.Filename
		att.Filesize = handler.Size
		db.Create(&att)

		// save file
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			return &HTTPError{err, "cant read file", 500}
		}
		err = ioutil.WriteFile(filepath.Join(config.mediaDir, "temp", sid.String(), att.ID.String()), fileBytes, os.ModePerm)
		if err != nil {
			db.Rollback()
			return &HTTPError{err, "cant save file", 500}
		}
		db.Commit()
	}

	return SendJSON(w, att)
}

func DownloadZip(w http.ResponseWriter, r *http.Request) *HTTPError {
	fmt.Println("DownloadZip")

	vars := mux.Vars(r)
	shareID, err := uuid.Parse(vars["id"])
	if err != nil {
		return &HTTPError{err, "invalid URL param", 400}
	}

	share, er := RetrieveShare(shareID, true)
	if er != nil {
		return er
	}

	zipWriter := zip.NewWriter(w)
	for _, file := range share.Attachments {
		filePath := filepath.Join(config.mediaDir, "data", file.ShareID.String(), file.ID.String())

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

/////////////////////////////////
////////// functions ////////////
/////////////////////////////////
func RetrieveShare(shareID uuid.UUID, withAtt bool) (*Share, *HTTPError) { // TODO move somewhere else
	db, err := GetDatabase()
	if err != nil {
		return nil, &HTTPError{err, "Can't get database", 500}
	}

	var share Share
	if withAtt == true {
		err = db.Preload("Attachments").Where("ID = ?", shareID).First(&share).Error
	} else {
		err = db.Where("ID = ?", shareID).First(&share).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &HTTPError{err, "Record not found", 404}
	}
	if err != nil {
		return nil, &HTTPError{err, "Can't fetch data", 500}
	}
	if share.IsTemporary == true {
		return nil, &HTTPError{errors.New("share is not finalized"), "Share is not finalized", 403}
	}
	return &share, nil
}

func ConfigureRoutes() {
	router := mux.NewRouter().StrictSlash(true)
	handler := cors.Default().Handler(router)

	router.Handle("/shares", endpointREST(AllShares)).Methods("GET")
	router.Handle("/shares", endpointREST(OpenShare)).Methods("POST")

	router.Handle("/share/{id}", endpointREST(GetShare)).Methods("GET")
	router.Handle("/share/{id}", endpointREST(CloseShare)).Methods("POST")

	router.Handle("/share/{id}/attachments", endpointREST(UploadAttachment)).Methods("POST")

	router.Handle("/share/{id}/attachment/{att}", endpointREST(DownloadFile)).Methods("GET")
	router.Handle("/share/{id}/zip", endpointREST(DownloadZip)).Methods("GET")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.port), handler))
}

func SendJSON(w http.ResponseWriter, res interface{}) *HTTPError {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return &HTTPError{err, "Can't encode data", 500}
	}
	return nil
}
