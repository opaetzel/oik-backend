package main

import (
	"encoding/json"
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
	w.WriteHeader(http.StatusOK)
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
