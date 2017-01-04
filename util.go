package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func internalError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	apiErr := jsonErr{Code: http.StatusInternalServerError, Message: "Internal server error. See log for details."}
	log.Println(err)
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		panic(err)
	}
}

func notFoundError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	apiErr := jsonErr{Code: http.StatusNotFound, Message: "Not found."}
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		panic(err)
	}
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	apiErr := jsonErr{Code: http.StatusUnauthorized, Message: "You are not authorized to see this data."}
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		panic(err)
	}
}

func notParsable(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(422)
	log.Println(err)
	apiErr := jsonErr{Code: 422, Message: "Error parsing input. See log for details."}
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		panic(err)
	}
}

func readBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 4194304))
	if err != nil {
		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		return nil, err
	}
	return body, nil
}

func notAcceptable(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotAcceptable)
	apiErr := jsonErr{Code: http.StatusNotAcceptable, Message: "Content-Type header not set or not acceptable"}
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		panic(err)
	}
}

func sendImage(w http.ResponseWriter, r *http.Request, imagePath string) {
	if _, err := os.Stat(imagePath); err == nil {
		file, err := os.Open(imagePath)
		if err != nil {
			internalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "image/"+imagePath[:len(imagePath)-3])
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
	} else {
		internalError(w, r, err)
	}
}

func sendRotateImage(w http.ResponseWriter, r *http.Request, imageFolder string, number int) {
	if _, err := os.Stat(imageFolder); err == nil {
		files, err := ioutil.ReadDir(imageFolder)
		if err != nil {
			internalError(w, r, err)
			return
		}
		if len(files)-1 < number {
			notFoundError(w, r)
			return
		}
		imagePath := filepath.Join(imageFolder, files[number].Name())
		sendImage(w, r, imagePath)
	} else {
		internalError(w, r, err)
	}
}
