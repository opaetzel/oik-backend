package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func registerRoute(router *mux.Router, route Route, handler http.Handler) {
	router.
		PathPrefix("/api/").
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name).
		Handler(handler)
}

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	//register public routes, no middleware needed
	for _, route := range publicRoutes {
		var handler http.Handler
		handler = route.Handler
		registerRoute(router, route, handler)
	}
	//register auth routes, only need to be logged in
	for _, route := range authRoutes {
		var handler http.Handler
		handler = jwtMiddleware.Handler(route.Handler)
		registerRoute(router, route, handler)
	}
	//register admin routes, need to be admin here. Maybe write another own middleware...
	for _, route := range adminRoutes {
		var handler http.Handler
		handler = jwtMiddleware.Handler(NewRequireRole(route.Handler, "admin"))
		registerRoute(router, route, handler)
	}
	fs := http.Dir("static/")
	fileHandler := http.FileServer(fs)
	router.PathPrefix("/").Handler(fileHandler)

	return router
}
