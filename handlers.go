package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var AllUnits = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if units, err := GetAllUnits(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apiErr := jsonErr{Code: http.StatusInternalServerError, Message: "Could not receive units. See log for details."}
		log.Println(err)
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			panic(err)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(units); err != nil {
			panic(err)
		}
	}

})

var PageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if pageId, err := strconv.Atoi(vars["pageId"]); err != nil {
		w.WriteHeader(422)
		apiErr := jsonErr{Code: http.StatusBadRequest, Message: "Could not parse pageId"}
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			panic(err)
		}
	} else {
		if page, err := GetPageById(pageId); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			apiErr := jsonErr{Code: http.StatusInternalServerError, Message: "Error getting page from DB. See log for details."}
			log.Println(err)
			if err := json.NewEncoder(w).Encode(apiErr); err != nil {
				panic(err)
			}
		} else {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(page); err != nil {
				panic(err)
			}
		}
	}
})

var PageCreate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var page Page
	//Set limit to 4MB, maybe make configurable later
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 4194304))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &page); err != nil {
		w.WriteHeader(422)
		log.Println(err)
		apiErr := jsonErr{Code: 422, Message: "Error parsing input. See log for details."}
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			panic(err)
		}
	}
	if err := InsertPage(page); err != nil {
		apiErr := jsonErr{Code: http.StatusInternalServerError, Message: "Error inserting to DB"}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			log.Println(err)
		}
	} else {
		w.WriteHeader(http.StatusCreated)
	}
})

//Remove this Handler later, or better replace by Login handler

var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	claims["groups"] = []string{"admin"}
	claims["name"] = "Felix Klein"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token.Claims = claims

	tokenString, _ := token.SignedString(mySigningKey)

	w.Write([]byte(tokenString))
})
