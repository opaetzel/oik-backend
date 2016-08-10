package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func AllUnits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if units, err := GetAllUnits(); err != nil {
		apiErr := jsonErr{Code: http.StatusInternalServerError, Message: "Could not receive units. See log for details."}
		log.Println(err)
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			panic(err)
		}
	} else {
		if err := json.NewEncoder(w).Encode(units); err != nil {
			panic(err)
		}
	}

}

func PageById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(422)
	if pageId, err := strconv.Atoi(vars["pageId"]); err != nil {
		apiErr := jsonErr{Code: http.StatusBadRequest, Message: "Could not parse pageId"}
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			panic(err)
		}
	} else {
		if page, err := GetPageById(pageId); err != nil {
			apiErr := jsonErr{Code: http.StatusInternalServerError, Message: "Error getting page from DB. See log for details."}
			log.Println(err)
			if err := json.NewEncoder(w).Encode(apiErr); err != nil {
				panic(err)
			}
		} else {
			if err := json.NewEncoder(w).Encode(page); err != nil {
				panic(err)
			}
		}
	}
}

func PageCreate(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422)
		log.Println(err)
		apiErr := jsonErr{Code: http.StatusBadRequest, Message: "Error parsing input. See log for details."}
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			panic(err)
		}
	}
	if err := InsertPage(page); err != nil {
		apiErr := jsonErr{Code: http.StatusInternalServerError, Message: "Error inserting to DB"}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			log.Println(err)
		}
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
	}
}
