package main

import "net/http"

type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

type Routes []Route

var editorRoutes = Routes{
	Route{
		"UnitCreate",
		"POST",
		"/units",
		UnitCreate,
	},
	Route{
		"UserUpdateUnit",
		"PUT",
		"/units/{unitId}",
		UpdateUnit,
	},
	Route{
		"PageCreate",
		"POST",
		"/pages",
		PageCreate,
	},
	Route{
		"PageUpdate",
		"PUT",
		"/pages/{pageId}",
		UserUpdatePage,
	},
	Route{
		"DeleteRow",
		"DELETE",
		"/rows/{rowId}",
		DeleteRow,
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
		UploadOrUpdateImage,
	},
	Route{
		"CreateRotateImage",
		"POST",
		"/rotateImages",
		CreateRotateImage,
	},
	/*	Route{
		"UpdateRotateImages",
		"PUT",
		"/rotateImages/{rotateImageId}",
		UpdateRotateImage,
	},*/
	Route{
		"UploadRotateImages",
		"PUT",
		"/upload-rotate-image/{rotateImageId}",
		UploadRotateImage,
	},
}

var authRoutes = Routes{
	Route{
		"GetUsers",
		"GET",
		"/users",
		AllUsers,
	},
	Route{
		"UpdateUser",
		"PUT",
		"/users/{userId}",
		UpdateUser,
	},
}

var publicRoutes = Routes{
	Route{
		"PublishedUnits",
		"GET",
		"/units",
		Units,
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
		"UserById",
		"GET",
		"/users/{userId}",
		UserById,
	},
	Route{
		"ImageById",
		"GET",
		"/images/{imageId}",
		ImageJSONById,
	},
	Route{
		"GetImageById",
		"GET",
		"/get-image/{imageId}",
		ImageById,
	},
	Route{
		"GetRotateImageObj",
		"GET",
		"/rotateImages/{rotateImageId}",
		RotateImageById,
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
		"/newusers",
		RegisterHandler,
	},
}
