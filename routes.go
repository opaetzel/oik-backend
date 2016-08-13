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
		"/admin/allunits",
		AllUnits,
	},
	Route{
		"AdminPageById",
		"GET",
		"/admin/pages/{pageId}",
		AdminPageById,
	},
	Route{
		"AdminUpdateUnit",
		"PUT",
		"/admin/units/{unitId}",
		AdminUpdateUnit,
	},
}

var authRoutes = Routes{
	Route{
		"UnitCreate",
		"POST",
		"/users/{userId}/units",
		UnitCreate,
	},
	Route{
		"UserUpdateUnit",
		"PUT",
		"/users/{userId}/units/{unitId}",
		UserUpdateUnit,
	},
	Route{
		"PageCreate",
		"POST",
		"/users/{userId}/units/{unitId}/pages",
		PageCreate,
	},
	Route{
		"PageUpdate",
		"PUT",
		"/users/{userId}/units/{unitId}/pages/{pageId}",
		UserUpdatePage,
	},
	Route{
		"UserUnits",
		"GET",
		"/users/{userId}/units",
		UserUnits,
	},
	Route{
		"UserPageById",
		"GET",
		"/users/{userId}/units/{unitId}/pages/{pageId}",
		UserPageById,
	},
	Route{
		"CreateImage",
		"POST",
		"/users/{userId}/units/{unitId}/images",
		CreateImage,
	},
	Route{
		"UploadImage",
		"PUT",
		"/users/{userId}/units/{unitId}/images",
		UploadImage,
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
