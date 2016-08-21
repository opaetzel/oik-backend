package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

/*
var AllUnits = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if units, err := GetAllUnits(); err != nil {
		internalError(w, r, err)
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"units": units}); err != nil {
			panic(err)
		}
	}
})
*/
var UnitById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	unitIdStr := mux.Vars(r)["unitId"]
	unitId, err := strconv.Atoi(unitIdStr)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	if unit, err := GetUnit(unitId); err != nil {
		internalError(w, r, err)
		return
	} else {
		if !unit.Published {
			user, err := getUserFromRequest(r)
			if err != nil {
				log.Println(err)
				unauthorized(w, r)
				return
			}
			if user.ID != unit.UserId && !user.isInGroup("admin") {
				unauthorized(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"unit": unit}); err != nil {
			panic(err)
		}
	}
})

var UserById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if userId, err := strconv.Atoi(mux.Vars(r)["userId"]); err != nil {
		notParsable(w, r, err)
		return
	} else {
		user, err := GetUserById(userId)
		if err != nil {
			notFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"user": user}); err != nil {
			panic(err)
		}
	}
})

var PublishedUnits = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if units, err := GetPublishedUnits(); err != nil {
		internalError(w, r, err)
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"units": units}); err != nil {
			panic(err)
		}
	}
})

var UserUpdateUnit = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
	}
	var unit Unit
	if err := json.Unmarshal(body, &unit); err != nil {
		notParsable(w, r, err)
		return
	}
	vars := mux.Vars(r)
	unitId, err := strconv.Atoi(vars["unitId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	if unit.ID != unitId {
		notParsable(w, r, err)
		return
	}
	if id, err := getUserId(r); err != nil {
		notParsable(w, r, err)
		return
	} else if id == unit.UserId {
		UpdateUnitUser(unit)
	} else {
		unauthorized(w, r)
	}
})

var PageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if pageId, err := strconv.Atoi(vars["pageId"]); err != nil {
		notParsable(w, r, err)
		return
	} else {
		if page, err := GetPageById(pageId); err != nil {
			internalError(w, r, err)
		} else {
			if !page.published {
				user, err := getUserFromRequest(r)
				if err != nil {
					log.Println(err)
					unauthorized(w, r)
					return
				}
				if user.ID != page.userId && !user.isInGroup("admin") {
					unauthorized(w, r)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"page": page}); err != nil {
				panic(err)
			}
		}
	}
})

var PageCreate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var page Page
	//Set limit to 4MB, maybe make configurable later
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if err := json.Unmarshal(body, &page); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := InsertPage(page); err != nil {
		internalError(w, r, err)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
})

var UserUpdatePage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	pageId, err := strconv.Atoi(vars["pageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	dbUserId, err := GetPageOwner(pageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	id, err := getUserId(r)
	if err != nil {
		notParsable(w, r, err)
		return
	} else {
		if dbUserId != id {
			unauthorized(w, r)
			return
		}
		body, err := readBody(r)
		if err != nil {
			internalError(w, r, err)
			return
		}
		var page Page
		err = json.Unmarshal(body, &page)
		if err != nil {
			notParsable(w, r, err)
			return
		}
		err = UpdatePage(page)
		if err != nil {
			internalError(w, r, err)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
})

var LoginOptionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Acces-Control-Allow-Headers", "content-type")
	w.Header().Set("Acces-Control-Allow-Methods", "POST")
	//w.Header().Set("Acces-Control-Allow-Origin", "*")
})

var LoginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	loginFailed := func() {
		w.WriteHeader(http.StatusUnauthorized)
		apiErr := jsonErr{Code: http.StatusUnauthorized, Message: "Wrong username or password"}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"error": apiErr}); err != nil {
			panic(err)
		}
	}
	fmt.Println("bla")
	var login LoginStruct
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	fmt.Println("try to unmarshal")
	if err := json.Unmarshal(body, &login); err != nil {
		notParsable(w, r, err)
		return
	} else {
		fmt.Println("unmarshal success")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		user, err := GetUserByName(login.Username)
		if err != nil {
			fmt.Println("did not get user")
			fmt.Println(err)
			loginFailed()
			return
		}
		fmt.Println("got user")
		hash, err := HashPWWithSaltB64(login.Password, user.salt)
		if err != nil {
			loginFailed()
			return
		}
		log.Println("hashed pw")
		pwHashBytes, err := base64.StdEncoding.DecodeString(user.pwHash)
		if err != nil {
			internalError(w, r, err)
			return
		}
		if bytes.Equal(pwHashBytes, hash) && user.active {
			token := jwt.New(jwt.SigningMethodHS256)
			claims := make(jwt.MapClaims)

			claims["groups"] = user.Groups
			claims["name"] = user.Username
			claims["exp"] = time.Now().Add(time.Hour * 12).Unix()
			claims["uid"] = user.ID

			token.Claims = claims

			tokenString, _ := token.SignedString(mySigningKey)

			if err := json.NewEncoder(w).Encode(map[string]interface{}{"token": tokenString}); err != nil {
				internalError(w, r, err)
			}
		} else {
			loginFailed()
			return
		}
	}
})

var RegisterHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var login LoginStruct
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if err := json.Unmarshal(body, &login); err != nil {
		notParsable(w, r, err)
		return
	}
	salt, pwhash, err := HashNewPW(login.Password)
	if err != nil {
		internalError(w, r, err)
		return
	}
	b64hash := base64.StdEncoding.EncodeToString(pwhash)
	user := User{login.Username, []string{"student"}, nil, 0, salt, b64hash, false}
	if err := InsertUser(user); err != nil {
		internalError(w, r, err)
		return
	}
})

var UnitCreate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var unit Unit
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if err := json.Unmarshal(body, &unit); err != nil {
		notParsable(w, r, err)
		return
	} else {
		err := InsertUnit(unit)
		if err != nil {
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
})

var CreateImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var image Image
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if err := json.Unmarshal(body, &image); err != nil {
		notParsable(w, r, err)
		return
	} else {
		imageId, err := InsertImage(image)
		if err != nil {
			internalError(w, r, err)
			return
		}
		image.ID = imageId
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"image": image}); err != nil {
			panic(err)
		}
	}
})

var UploadImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(r)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	vars := mux.Vars(r)
	imageId, err := strconv.Atoi(vars["imageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	log.Println("try to get image owner for image", imageId)
	image, err := GetImageById(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if image.UserId != userId {
		unauthorized(w, r)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType == "" || (contentType != "image/jpeg" && contentType != "image/png") {
		log.Println(contentType)
		notAcceptable(w, r)
		return
	}
	extension := ".jpg"
	if contentType == "image/png" {
		extension = ".png"
	}
	imageDir := filepath.Join(conf.ImageStorage, strconv.Itoa(image.UserId))
	os.MkdirAll(imageDir, 0755)
	imagePath := filepath.Join(imageDir, strconv.Itoa(image.ID)+extension)
	outFile, err := os.Create(imagePath)
	if err != nil {
		internalError(w, r, err)
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, r.Body)
	if err != nil {
		internalError(w, r, err)
		return
	} else {
		err := UpdateImagePath(imageId, imagePath)
		if err != nil {
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
})

var RotateImageByIdAndNumber = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageId, err := strconv.Atoi(vars["imageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	number, err := strconv.Atoi(vars["number"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	published, imageFolder, err := GetRotateImagePublishedAndPath(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if !published {
		unauthorized(w, r)
		return
	}
	sendRotateImage(w, r, imageFolder, number)
})

var ImageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageId, err := strconv.Atoi(vars["imageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	image, err := GetImageById(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	//TODO: In frontend: send xhr with token on un-published images
	if !image.published {
		user, err := getUserFromRequest(r)
		if err != nil {
			log.Println(err)
			unauthorized(w, r)
			return
		}
		if user.ID != image.UserId && !stringInSlice("admin", user.Groups) {
			unauthorized(w, r)
			return
		}
	}
	sendImage(w, r, image.path)
})
