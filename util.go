package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
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

func getUserFromRequest(r *http.Request) (User, error) {
	userJWT := context.Get(r, "user")
	if userJWT == nil {
		return User{}, errors.New("no token in context")
	}
	claims, ok := userJWT.(*jwt.Token).Claims.(jwt.MapClaims)
	if ok {
		claimIdF, ok := claims["uid"].(float64)
		if !ok {
			return User{}, errors.New("could not cast uid to int")
		}
		claimId := int(claimIdF)
		claimGroups, ok := claims["groups"].([]interface{})
		if !ok {
			return User{}, errors.New("could not parse groups")
		}
		groups := make([]string, len(claimGroups))
		for i, gr := range claimGroups {
			group, ok := gr.(string)
			if !ok {
				return User{}, errors.New("could not parse groups")
			}
			groups[i] = group
		}
		name, ok := claims["name"].(string)
		if !ok {
			return User{}, errors.New("could not parse name")
		}
		return User{Username: name, Groups: groups, ID: claimId}, nil

	} else {
		return User{}, errors.New("could not read claims")
	}
}

func getUserId(r *http.Request) (int, error) {
	user := context.Get(r, "user")
	claims, ok := user.(*jwt.Token).Claims.(jwt.MapClaims)
	if ok {
		claimIdF, ok := claims["uid"].(float64)
		if !ok {
			return 0, errors.New("could not cast uid to int")
		}
		claimId := int(claimIdF)
		return claimId, nil
	} else {
		return 0, errors.New("could not read claims")
	}
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
