package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.Handler
		//		handler = Logger(handler, route.Name)

		router.
			PathPrefix("/api/").
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	fs := http.Dir("static/")
	fileHandler := http.FileServer(fs)
	router.PathPrefix("/").Handler(fileHandler)

	return router
}
