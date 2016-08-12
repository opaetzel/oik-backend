package main

import "net/http"

type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

type Routes []Route

var adminRoutes = Routes{
	Route{
		"AllUnits",
		"GET",
		"/allunits",
		AllUnits,
	},
	Route{
		"AdminPageById",
		"GET",
		"/allunits",
		AdminPageById,
	},
}

var authRoutes = Routes{
	Route{
		"UnitCreate",
		"POST",
		"/units",
		UnitCreate,
	},
	Route{
		"PageCreate",
		"POST",
		"/pages",
		PageCreate,
	},
	Route{
		"UserUnits",
		"GET",
		"/{userId}/units",
		UserUnits,
	},
	Route{
		"UserPageById",
		"GET",
		"/{userId}/page/{pageId}",
		UserPageById,
	},
}

var publicRoutes = Routes{
	Route{
		"PublishedUnits",
		"GET",
		"/units",
		PublishedUnits,
	},
	Route{
		"PageById",
		"GET",
		"/pages/{pageId}",
		PageById,
	},
	Route{
		"Login",
		"POST",
		"/login",
		LoginHandler,
	},
	Route{
		"Register",
		"POST",
		"/register",
		RegisterHandler,
	},
}
