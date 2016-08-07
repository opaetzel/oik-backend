package main

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"AllUnits",
		"GET",
		"/allunits",
		AllUnits,
	},
	Route{
		"PageById",
		"GET",
		"/pages/{pageId}",
		PageById,
	},
}
