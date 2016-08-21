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
		"/admin-pages/{pageId}",
		AdminPageById,
	},
	Route{
		"AdminUpdateUnit",
		"PATCH",
		"/admin-units/{unitId}",
		AdminUpdateUnit,
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
		"UserUpdateUnit",
		"PATCH",
		"/units/{unitId}",
		UserUpdateUnit,
	},
	Route{
		"PageCreate",
		"POST",
		"/pages",
		PageCreate,
	},
	Route{
		"PageUpdate",
		"PATCH",
		"/pages/{pageId}",
		UserUpdatePage,
	},
	Route{
		"UserUnits",
		"GET",
		"/user-units",
		UserUnits,
	},
	Route{
		"UserPageById",
		"GET",
		"/user-pages/{pageId}",
		UserPageById,
	},
	Route{
		"CreateImage",
		"POST",
		"/images",
		CreateImage,
	},
	Route{
		"UploadImage",
		"PUT",
		"/images/{imageId}",
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
		"UnitById",
		"GET",
		"/units/{unitId}",
		UnitById,
	},
	Route{
		"PageById",
		"GET",
		"/pages/{pageId}",
		PageById,
	},
	Route{
		"ImageById",
		"GET",
		"/get-image/{imageId}",
		ImageById,
	},
	Route{
		"RotateImageByIdAndNumber",
		"GET",
		"/get-rotate-image/{imageId}/{number}",
		RotateImageByIdAndNumber,
	},
	Route{
		"LoginOptions",
		"OPTIONS",
		"/login",
		LoginOptionsHandler,
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
