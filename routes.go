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
		"UnitDelete",
		"DELETE",
		"/units/{unitId}",
		DeleteUnit,
	},
}

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
		UpdatePage,
	},
	Route{
		"PageDelete",
		"DELETE",
		"/pages/{pageId}",
		DeletePage,
	},
	Route{
		"DeleteRow",
		"DELETE",
		"/rows/{rowId}",
		DeleteRow,
	},
	Route{
		"CreateErrorImage",
		"POST",
		"/errorImages",
		CreateErrorImage,
	},
	Route{
		"UploadErrorImage",
		"PUT",
		"/errorImages/{errorImageId}",
		UploadOrUpdateErrorImage,
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
		"GetPageResults",
		"GET",
		"/pageResults/{pageResultId}",
		GetPageResult,
	},
	Route{
		"InsertUnitResult",
		"POST",
		"/unitResults",
		InsertUnitResult,
	},
	Route{
		"UpdateUnitResult",
		"PUT",
		"/unitResults/{unitId}",
		UpdateUnitResult,
	},
	Route{
		"InsertPageResult",
		"POST",
		"/pageResults",
		InsertPageResult,
	},
	Route{
		"UpdatePageResult",
		"PUT",
		"/pageResults/{pageId}",
		UpdatePageResult,
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
		"NewPasswordRequest",
		"POST",
		"/newPasswordRequests",
		NewPasswordRequest,
	},
	Route{
		"GetUnitResults",
		"GET",
		"/unitResults/{unitId}",
		GetUnitResult,
	},
	Route{
		"PublishedUnits",
		"GET",
		"/units",
		Units,
	},
	Route{
		"AllErrorImages",
		"GET",
		"/errorImages",
		ErrorImages,
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
		"GetErrorImageById",
		"GET",
		"/get-error-image/{imageId}",
		ErrorImageById,
	},
	Route{
		"GetErrorImage",
		"GET",
		"/errorImages/{imageId}",
		ErrorImageJSONById,
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
