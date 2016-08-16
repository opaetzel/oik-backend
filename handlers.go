package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

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

var UserUnits = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userIdStr := mux.Vars(r)["userId"]
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	user := context.Get(r, "user")
	claims, ok := user.(*jwt.Token).Claims.(jwt.MapClaims)
	claimUserId, ok := claims["uid"].(int)
	if !ok {
		internalError(w, r, errors.New("claimUserId not string"))
		return
	}
	if claimUserId != userId {
		unauthorized(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if units, err := GetUserUnits(userId); err != nil {
		internalError(w, r, err)
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"units": units}); err != nil {
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

var AdminUpdateUnit = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	UpdateUnitAdmin(unit)
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
	if userOK, id, err := checkUserId(r); err != nil {
		notParsable(w, r, err)
		return
	} else if userOK && id == unit.UserId {
		UpdateUnitUser(unit)
	} else {
		unauthorized(w, r)
	}
})

var UserPageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	pageId, err := strconv.Atoi(vars["pageId"])
	if err != nil {
		notParsable(w, r, err)
	}
	userOK, id, err := checkUserId(r)
	if err != nil {
		notParsable(w, r, err)
	} else if userOK {
		if page, err := GetUserPageById(id, pageId); err != nil {
			internalError(w, r, err)
		} else {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"page": page}); err != nil {
				panic(err)
			}
		}

	} else {
		unauthorized(w, r)
		return
	}
})

var AdminPageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if pageId, err := strconv.Atoi(vars["pageId"]); err != nil {
		notParsable(w, r, err)
		return
	} else {
		if page, err := GetPageById(pageId); err != nil {
			internalError(w, r, err)
		} else {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"page": page}); err != nil {
				panic(err)
			}
		}
	}
})

var PageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if pageId, err := strconv.Atoi(vars["pageId"]); err != nil {
		notParsable(w, r, err)
		return
	} else {
		if page, err := GetPublicPageById(pageId); err != nil {
			internalError(w, r, err)
		} else {
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
	userOK, id, err := checkUserId(r)
	if err != nil {
		notParsable(w, r, err)
		return
	} else if userOK {
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
	} else {
		unauthorized(w, r)
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
		user, err := GetUser(login.Username)
		if err != nil {
			fmt.Println("did not get user")
			fmt.Println(err)
			loginFailed()
			return
		}
		fmt.Println("got user")
		hash, err := HashPWWithSaltB64(login.Password, user.Salt)
		if err != nil {
			loginFailed()
			return
		}
		log.Println("hashed pw")
		pwHashBytes, err := base64.StdEncoding.DecodeString(user.PWHash)
		if err != nil {
			internalError(w, r, err)
			return
		}
		log.Println(pwHashBytes, hash)
		if bytes.Equal(pwHashBytes, hash) && user.Active {
			token := jwt.New(jwt.SigningMethodHS256)
			claims := make(jwt.MapClaims)

			claims["groups"] = user.Groups
			claims["name"] = user.Username
			claims["exp"] = time.Now().Add(time.Hour * 12).Unix()
			claims["uid"] = user.ID

			token.Claims = claims

			tokenString, _ := token.SignedString(mySigningKey)

			w.Write([]byte(tokenString))
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
	user := User{login.Username, []string{"student"}, salt, b64hash, false, 0}
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
	ok, userId, err := checkUserId(r)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	if !ok {
		unauthorized(w, r)
		return
	}
	vars := mux.Vars(r)
	imageId, err := strconv.Atoi(vars["imageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	log.Println("try to get image owner for image", imageId)
	imageOwner, err := GetImageOwner(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if imageOwner != userId {
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
	imageDir := filepath.Join(conf.ImageStorage, strconv.Itoa(imageOwner))
	os.MkdirAll(imageDir, 0755)
	imagePath := filepath.Join(imageDir, strconv.Itoa(imageId)+extension)
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
		w.WriteHeader(http.StatusCreated)
	}
})
