package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
