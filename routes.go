package main

import "net/http"

type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

type Routes []Route

var routes = Routes{
	Route{
		"AllUnits",
		"GET",
		"/allunits",
		jwtMiddleware.Handler(AllUnits),
	},
	Route{
		"PageById",
		"GET",
		"/pages/{pageId}",
		PageById,
	},
	Route{
		"PageCreate",
		"POST",
		"/pages",
		PageCreate,
	},
	Route{
		"Login",
		"POST",
		"/login",
		LoginHandler,
	},
}
