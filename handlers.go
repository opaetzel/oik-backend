package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
		if err := json.NewEncoder(w).Encode(units); err != nil {
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
		if err := json.NewEncoder(w).Encode(units); err != nil {
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
		if err := json.NewEncoder(w).Encode(units); err != nil {
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
			if err := json.NewEncoder(w).Encode(page); err != nil {
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
			if err := json.NewEncoder(w).Encode(page); err != nil {
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
			if err := json.NewEncoder(w).Encode(page); err != nil {
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

var LoginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	loginFailed := func() {
		w.WriteHeader(http.StatusUnauthorized)
		apiErr := jsonErr{Code: http.StatusUnauthorized, Message: "Wrong username or password"}
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
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
		}
		w.WriteHeader(http.StatusCreated)
	}
})
