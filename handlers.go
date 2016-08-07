package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func AllUnits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	//TODO: get units from db, without content
	fmt.Fprint(w, "[]")
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
		//TODO: get from db, if not found: return err 404
		page := Page{"Sample page", nil, pageId}
		if err := json.NewEncoder(w).Encode(page); err != nil {
			panic(err)
		}
	}
}
