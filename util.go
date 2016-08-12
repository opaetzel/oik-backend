package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
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

func checkUserId(r *http.Request) (bool, int, error) {
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["userId"])
	if err != nil {
		return false, 0, err
	}
	user := context.Get(r, "user")
	claims, ok := user.(*jwt.Token).Claims.(jwt.MapClaims)
	if ok {
		claimId, ok := claims["uid"].(int)
		if !ok {
			return false, 0, errors.New("could not cast uid to int")
		}
		if claimId != userId {
			return false, 0, nil
		} else {
			return true, userId, nil
		}
	} else {
		return false, 0, errors.New("could not read claims")
	}
}
