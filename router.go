package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func registerRoute(router *mux.Router, route Route, handler http.Handler) {
	router.
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name).
		Handler(handler)
}

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	api := router.PathPrefix("/api/").Subrouter()
	//register public routes, no middleware needed
	for _, route := range publicRoutes {
		var handler http.Handler
		handler = route.Handler
		registerRoute(api, route, handler)
	}
	//register auth routes, only need to be logged in
	for _, route := range authRoutes {
		var handler http.Handler
		handler = jwtMiddleware.Handler(route.Handler)
		registerRoute(api, route, handler)
	}
	//register admin routes, need to be admin here.
	for _, route := range adminRoutes {
		var handler http.Handler
		handler = jwtMiddleware.Handler(NewRequireRole(route.Handler, "admin"))
		registerRoute(api, route, handler)
	}
	fs := http.Dir(conf.StaticFolder)
	fileHandler := http.FileServer(fs)
	router.PathPrefix("/assets/").Handler(fileHandler)
	router.Path("/crossdomain.xml").Handler(fileHandler)
	router.Path("/robots.txt").Handler(fileHandler)
	router.NotFoundHandler = http.HandlerFunc(notFound)

	return router
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
