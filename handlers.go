package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	goimg "image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
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
			if user.ID != unit.UserId && !user.isInGroup("admin") && !user.isInGroup("editor") {
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
			log.Println(err)
			notFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"user": user}); err != nil {
			panic(err)
		}
	}
})

var Units = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	publishedFilter := r.URL.Query().Get("filter[published]")
	if len(publishedFilter) == 0 || publishedFilter == "true" {
		units, err := GetPublishedUnits()
		if err != nil {
			internalError(w, r, err)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"units": units}); err != nil {
			panic(err)
		}
	} else {
		user, err := getUserFromRequest(r)
		if err != nil {
			unauthorized(w, r)
			return
		}
		if !user.isInGroup("admin") && !user.isInGroup("editor") {
			unauthorized(w, r)
			return
		}
		units, err := GetUnPublishedUnits()
		if err != nil {
			internalError(w, r, err)
			return
		}
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

var UpdateUnit = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	var unit Unit
	if err := json.Unmarshal(*objmap["unit"], &unit); err != nil {
		notParsable(w, r, err)
		return
	}
	vars := mux.Vars(r)
	unitId, err := strconv.Atoi(vars["unitId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	unit.ID = unitId
	if user, err := getUserFromRequest(r); err != nil {
		notParsable(w, r, err)
		return
	} else if user.isInGroup("admin") {
		err := UpdateUnitAdmin(unit)
		if err != nil {
			internalError(w, r, err)
			return
		}
		if _, err := w.Write([]byte("{}")); err != nil {
			panic(err)
		}
	} else if user.ID == unit.UserId {
		err := UpdateUnitUser(unit)
		if err != nil {
			internalError(w, r, err)
			return
		}
		if _, err := w.Write([]byte("{}")); err != nil {
			panic(err)
		}
	} else {
		unauthorized(w, r)
	}
})

var PageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	user, _ := getUserFromRequest(r)
	if pageId, err := strconv.Atoi(vars["pageId"]); err != nil {
		notParsable(w, r, err)
		return
	} else {
		var page Page
		var err error
		if user.ID > 0 {
			page, err = GetPageWithResultsById(pageId, user.ID)
		} else {
			page, err = GetPageById(pageId)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				notFoundError(w, r)
				return
			} else {
				internalError(w, r, err)
				return
			}
		}
		if !page.published {
			if err != nil {
				log.Println(err)
				unauthorized(w, r)
				return
			}
			if user.ID != page.userId && !user.isInGroup("admin") && !user.isInGroup("editor") {
				unauthorized(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"page": page}); err != nil {
			panic(err)
		}
	}
})

var RotateImageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if imageId, err := strconv.Atoi(vars["rotateImageId"]); err != nil {
		notParsable(w, r, err)
		return
	} else {
		if image, err := GetRotateImageById(imageId); err != nil {
			internalError(w, r, err)
		} else {
			if !image.published {
				user, err := getUserFromRequest(r)
				if err != nil {
					log.Println(err)
					unauthorized(w, r)
					return
				}
				if user.ID != image.UserId && !user.isInGroup("admin") {
					unauthorized(w, r)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"rotateImage": image}); err != nil {
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
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["page"], &page); err != nil {
		notParsable(w, r, err)
		return
	}
	log.Println(page)
	if insertedPage, err := InsertPage(page); err != nil {
		internalError(w, r, err)
	} else {
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"page": insertedPage}); err != nil {
			panic(err)
		}
	}
})

var DeletePage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	user, err := getUserFromRequest(r)
	if err != nil {
		notParsable(w, r, err)
		return
	} else {
		if dbUserId != user.ID && !user.isInGroup("admin") {
			unauthorized(w, r)
			return
		}
		err := DbDeletePage(pageId)
		if err != nil {
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
})

var UpdatePage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	user, err := getUserFromRequest(r)
	if err != nil {
		notParsable(w, r, err)
		return
	} else {
		if dbUserId != user.ID && !user.isInGroup("admin") {
			unauthorized(w, r)
			return
		}
		body, err := readBody(r)
		if err != nil {
			internalError(w, r, err)
			return
		}
		var objmap map[string]*json.RawMessage
		if err := json.Unmarshal(body, &objmap); err != nil {
			notParsable(w, r, err)
			return
		}
		var page Page
		err = json.Unmarshal(*objmap["page"], &page)
		if err != nil {
			notParsable(w, r, err)
			return
		}
		page.ID = pageId
		page, err = DbUpdatePage(page)
		if err != nil {
			internalError(w, r, err)
		} else {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"page": page}); err != nil {
				panic(err)
			}
		}
	}
})

var LoginOptionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Acces-Control-Allow-Headers", "content-type")
	w.Header().Set("Acces-Control-Allow-Methods", "POST")
	//w.Header().Set("Acces-Control-Allow-Origin", "*")
})

func loginWithToken(w http.ResponseWriter, r *http.Request, token string) {
	parser := jwt.Parser{[]string{jwt.SigningMethodHS256.Name}, false, false}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	}
	_, err := parser.Parse(token, keyFunc)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"token": token}); err != nil {
		internalError(w, r, err)
	}
}

var LoginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	loginFailed := func() {
		w.WriteHeader(http.StatusUnauthorized)
		apiErr := jsonErr{Code: http.StatusUnauthorized, Message: "Wrong username or password"}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"error": apiErr}); err != nil {
			panic(err)
		}
	}
	log.Println("bla")
	var login LoginStruct
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	log.Println("try to unmarshal")
	if err := json.Unmarshal(body, &login); err != nil {
		notParsable(w, r, err)
		return
	} else {
		log.Println("unmarshal success")
		if len(login.Username) == 0 {
			loginWithToken(w, r, login.Password)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		user, err := GetUserByName(login.Username)
		if err != nil {
			log.Println("did not get user")
			log.Println(err)
			loginFailed()
			return
		}
		log.Println("got user")
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
		if bytes.Equal(pwHashBytes, hash) && user.Active {
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
			log.Println("user active:", user.Active)
			log.Println(pwHashBytes)
			log.Println(hash)
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
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["newuser"], &login); err != nil {
		notParsable(w, r, err)
		return
	}
	if userInDb, err := IsUsernameInDb(login.Username); err != nil {
		internalError(w, r, err)
		return
	} else if userInDb {
		w.WriteHeader(http.StatusConflict)
		jsonError := jsonErr{http.StatusConflict, "Username already exists"}
		if err := json.NewEncoder(w).Encode(jsonError); err != nil {
			panic(err)
		}
		return
	}
	salt, pwhash, err := HashNewPW(login.Password)
	if err != nil {
		internalError(w, r, err)
		return
	}
	b64hash := base64.StdEncoding.EncodeToString(pwhash)
	mailHash, err := HashPWWithSaltB64(login.Email, salt)
	b64MailHash := base64.StdEncoding.EncodeToString(mailHash)
	user := User{login.Username, []string{"student"}, nil, 0, salt, b64hash, false, b64MailHash, 0, 0, "", nil, nil}
	if userId, err := InsertUser(user); err != nil {
		internalError(w, r, err)
		return
	} else {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := make(jwt.MapClaims)

		claims["groups"] = make([]string, 0)
		claims["name"] = user.Username
		claims["exp"] = time.Now().Add(time.Hour * 12).Unix()
		claims["uid"] = userId

		token.Claims = claims

		tokenString, _ := token.SignedString(mySigningKey)
		if err := sendMail(login.Email, conf.AppUrl+"confirm-mail/"+tokenString, "Registrierung Objekte im Kreuzverhör", registerMailTemplate); err != nil {
			log.Printf("Error sending mail\n")
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("{}")); err != nil {
			panic(err)
		}
	}
})

var UpdateUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["userId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	var user User
	body, err := readBody(r)
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["user"], &user); err != nil {
		notParsable(w, r, err)
		return
	}
	user.ID = userId
	claims, err := GetJWTClaims(r)
	if err != nil {
		unauthorized(w, r)
		return
	}
	claimIdF, ok := claims["uid"].(float64)
	if !ok {
		internalError(w, r, errors.New("could not cast uid"))
		return
	}
	claimId := int(claimIdF)
	if user.ID == claimId {
		if len(user.NewPw) != 0 {
			dbUser, err := GetUserByName(user.Username)
			pwhash, err := HashPWWithSaltB64(user.NewPw, dbUser.salt)
			log.Println(dbUser.salt)
			if err != nil {
				internalError(w, r, err)
				return
			}
			b64hash := base64.StdEncoding.EncodeToString(pwhash)
			log.Println(b64hash)
			user.pwHash = b64hash
			user.salt = dbUser.salt
			UpdateUserPW(user)
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"user": user}); err != nil {
				internalError(w, r, err)
			}
			return
		}
		UserUpdateUser(user)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"user": user}); err != nil {
			internalError(w, r, err)
		}
	} else {
		groups, err := GetClaimGroups(claims)
		if err != nil {
			internalError(w, r, err)
			return
		}
		if stringInSlice("admin", groups) {
			AdminUpdateUser(user)
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"user": user}); err != nil {
				internalError(w, r, err)
			}
		} else {
			unauthorized(w, r)
			return
		}
	}
})

var UnitCreate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var unit Unit
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["unit"], &unit); err != nil {
		notParsable(w, r, err)
		return
	} else {
		id, err := InsertUnit(unit)
		if err != nil {
			internalError(w, r, err)
			return
		}
		unit.ID = id
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"unit": unit}); err != nil {
			panic(err)
		}
	}
})

var CreateImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var image Image
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["image"], &image); err != nil {
		notParsable(w, r, err)
		return
	} else {
		log.Println(image)
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

var CreateErrorImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	var errorImage ErrorImage
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["errorImage"], &errorImage); err != nil {
		notParsable(w, r, err)
		return
	} else {
		log.Println(errorImage)
		errorImage.UserId = user.ID
		errorImage, err := InsertErrorImage(errorImage)
		if err != nil {
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"errorImage": errorImage}); err != nil {
			panic(err)
		}
	}
})

var UploadOrUpdateErrorImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	vars := mux.Vars(r)
	errorImageId, err := strconv.Atoi(vars["errorImageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	log.Println("try to get errorImage owner for errorImage", errorImageId)
	errorImage, err := GetErrorImageById(errorImageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if errorImage.UserId != user.ID {
		log.Println("userid != errorImage.userid", user.ID, "!=", errorImage.UserId)
		unauthorized(w, r)
		return
	}
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		//no formfile. Try to update errorImage.
		log.Println("does not seem to be a formfile")
		var updateErrorImage ErrorImage
		body, err := readBody(r)
		if err != nil {
			internalError(w, r, err)
			return
		}
		var objmap map[string]*json.RawMessage
		if err := json.Unmarshal(body, &objmap); err != nil {
			notParsable(w, r, err)
			return
		}
		if err := json.Unmarshal(*objmap["errorImage"], &updateErrorImage); err != nil {
			notParsable(w, r, err)
			return
		}
		updateErrorImage.ID = errorImageId
		updateErrorImage, err = UpdateErrorImage(updateErrorImage)
		if err != nil {
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"errorImage": updateErrorImage}); err != nil {
			panic(err)
		}
		return
	}
	defer file.Close()
	filename := (*header).Filename
	splitFilename := strings.Split(filename, ".")
	if len(splitFilename) < 2 {
		log.Printf("can not accept filename: %s\n", filename)
		notAcceptable(w, r)
		return
	}
	extension := "." + strings.ToLower(splitFilename[len(splitFilename)-1])
	if extension != ".png" && extension != ".jpeg" && extension != ".jpg" {
		log.Printf("can not accept filename: %s\n", filename)
		notAcceptable(w, r)
		return
	}
	errorImageDir := filepath.Join(conf.ImageStorage, strconv.Itoa(errorImage.UserId))
	os.MkdirAll(errorImageDir, 0755)
	errorImagePath := filepath.Join(errorImageDir, strconv.Itoa(errorImage.ID)+extension)
	outFile, err := os.Create(errorImagePath)
	if err != nil {
		internalError(w, r, err)
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, file)
	if err != nil {
		internalError(w, r, err)
		return
	}
	idx := strings.LastIndex(errorImagePath, ".")
	smallPath := errorImagePath[0:idx] + "_small" + errorImagePath[idx:len(errorImagePath)]
	_, err = file.Seek(0, 0)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var img goimg.Image
	if strings.Contains(extension, "jpg") || strings.Contains(extension, "jpeg") {
		img, err = jpeg.Decode(file)
		if err != nil {
			internalError(w, r, err)
			return
		}
	} else if strings.Contains(extension, "png") {
		img, err = png.Decode(file)
		if err != nil {
			internalError(w, r, err)
			return
		}
	}

	m := resize.Thumbnail(550, 330, img, resize.Lanczos3)

	smallOutFile, err := os.Create(smallPath)
	if err != nil {
		internalError(w, r, err)
		return
	}
	defer smallOutFile.Close()
	jpeg.Encode(smallOutFile, m, nil)

	err = UpdateErrorImagePath(errorImageId, errorImagePath)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)

})

var UploadOrUpdateImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
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
	if image.UserId != user.ID {
		log.Println("userid != image.userid", user.ID, "!=", image.UserId)
		unauthorized(w, r)
		return
	}
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		//no formfile. Try to update image.
		log.Println("does not seem to be a formfile")
		var updateImage Image
		body, err := readBody(r)
		if err != nil {
			internalError(w, r, err)
			return
		}
		var objmap map[string]*json.RawMessage
		if err := json.Unmarshal(body, &objmap); err != nil {
			notParsable(w, r, err)
			return
		}
		if err := json.Unmarshal(*objmap["image"], &updateImage); err != nil {
			notParsable(w, r, err)
			return
		}
		updateImage.ID = imageId
		err = UpdateImageUser(updateImage)
		if err != nil {
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"image": updateImage}); err != nil {
			panic(err)
		}
		return
	}
	defer file.Close()
	filename := (*header).Filename
	splitFilename := strings.Split(filename, ".")
	if len(splitFilename) < 2 {
		log.Printf("can not accept filename: %s\n", filename)
		notAcceptable(w, r)
		return
	}
	extension := "." + strings.ToLower(splitFilename[len(splitFilename)-1])
	if extension != ".png" && extension != ".jpeg" && extension != ".jpg" {
		log.Printf("can not accept filename: %s\n", filename)
		notAcceptable(w, r)
		return
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
	_, err = io.Copy(outFile, file)
	if err != nil {
		internalError(w, r, err)
		return
	}
	idx := strings.LastIndex(imagePath, ".")
	smallPath := imagePath[0:idx] + "_small" + imagePath[idx:len(imagePath)]
	_, err = file.Seek(0, 0)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var img goimg.Image
	if strings.Contains(extension, "jpg") || strings.Contains(extension, "jpeg") {
		img, err = jpeg.Decode(file)
		if err != nil {
			internalError(w, r, err)
			return
		}
	} else if strings.Contains(extension, "png") {
		img, err = png.Decode(file)
		if err != nil {
			internalError(w, r, err)
			return
		}
	}

	m := resize.Thumbnail(550, 330, img, resize.Lanczos3)

	smallOutFile, err := os.Create(smallPath)
	if err != nil {
		internalError(w, r, err)
		return
	}
	defer smallOutFile.Close()
	jpeg.Encode(smallOutFile, m, nil)

	err = UpdateImagePath(imageId, imagePath)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)

})

var CreateRotateImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var image RotateImage
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["rotateImage"], &image); err != nil {
		notParsable(w, r, err)
		return
	} else {
		log.Println(image)
		imageId, err := InsertRotateImage(image)
		if err != nil {
			internalError(w, r, err)
			return
		}
		image.ID = imageId
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"rotateImage": image}); err != nil {
			panic(err)
		}
	}
})

/*
Awaits an tar.gz as FormFile. Extracts and saves image files from it.
*/
var UploadRotateImage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	log.Println("in UploadRotateImage")
	user, err := getUserFromRequest(r)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	vars := mux.Vars(r)
	log.Println("got vars: ", vars["rotateImageId"])
	imageId, err := strconv.Atoi(vars["rotateImageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	log.Println("try to get image owner for image", imageId)
	image, err := GetRotateImageById(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if image.UserId != user.ID {
		log.Println("userid != image.userid", user.ID, "!=", image.UserId)
		unauthorized(w, r)
		return
	}
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		notParsable(w, r, err)
		return
	}
	defer file.Close()
	//gunzip
	gReader, err := gzip.NewReader(file)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	defer gReader.Close()
	//read tar container
	tarReader := tar.NewReader(gReader)
	if err != nil {
		notParsable(w, r, err)
		return
	}
	imageDir := filepath.Join(conf.ImageStorage, strconv.Itoa(image.UserId), strconv.Itoa(image.ID))
	os.MkdirAll(imageDir, 0755)
	imageCount := 0
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			log.Printf("File not parsable\n")
			notParsable(w, r, err)
			return
		}
		splitFilename := strings.Split(hdr.Name, ".")
		if len(splitFilename) < 2 {
			log.Printf("can not accept filename: %s\n", hdr.Name)
			notAcceptable(w, r)
			return
		}
		extension := "." + strings.ToLower(splitFilename[len(splitFilename)-1])
		if extension != ".png" && extension != ".jpeg" && extension != ".jpg" {
			log.Printf("can not accept filename: %s\n", hdr.Name)
			notAcceptable(w, r)
			return
		}
		imagePath := filepath.Join(imageDir, fmt.Sprintf("%04d", imageCount)+extension)
		outFile, err := os.Create(imagePath)
		if err != nil {
			internalError(w, r, err)
			return
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			internalError(w, r, err)
			return
		}
		imageCount++
	}
	if err := SetRotateImagePathAndNum(image.ID, imageDir, imageCount); err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
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
	image, err := GetRotateImageById(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if !image.published {
		user, err := getUserFromRequest(r)
		if err != nil {
			log.Println("could not get user from request")
			unauthorized(w, r)
		}
		if user.ID != image.UserId && !stringInSlice("admin", user.Groups) {
			log.Println("rotateImage.userId != userid and user is not admin")
			unauthorized(w, r)
			return
		}
	}
	sendRotateImage(w, r, image.basepath, number)
})

var ErrorImageJSONById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageId, err := strconv.Atoi(vars["imageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	image, err := GetErrorImageById(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	/*
		if !image.published {
			user, err := getUserFromRequest(r)
			if err != nil {
				log.Println(err)
				unauthorized(w, r)
				return
			}
			if user.ID != image.UserId && !user.isInGroup("admin") && !user.isInGroup("editor") {
				unauthorized(w, r)
				return
			}
		}
	*/
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"errorImage": image}); err != nil {
		panic(err)
	}
})

var ImageJSONById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if !image.published {
		user, err := getUserFromRequest(r)
		if err != nil {
			log.Println(err)
			unauthorized(w, r)
			return
		}
		if user.ID != image.UserId && !user.isInGroup("admin") && !user.isInGroup("editor") {
			unauthorized(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"image": image}); err != nil {
		panic(err)
	}
})

var ImageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sendLargeIm := false
	sizeStr := r.URL.Query().Get("size")
	if sizeStr == "full" {
		sendLargeIm = true
	}
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
	if !image.published {
		user, err := getUserFromRequest(r)
		if err != nil {
			log.Println(err)
			unauthorized(w, r)
			return
		}
		if user.ID != image.UserId && !user.isInGroup("admin") && !user.isInGroup("editor") {
			unauthorized(w, r)
			return
		}
	}
	if sendLargeIm {
		sendImage(w, r, image.path)
	} else {
		idx := strings.LastIndex(image.path, ".")
		smallPath := image.path[0:idx] + "_small" + image.path[idx:len(image.path)]
		sendImage(w, r, smallPath)
	}
})

var ErrorImageById = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageId, err := strconv.Atoi(vars["imageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	image, err := GetErrorImageById(imageId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	/*
		if !image.published {
			user, err := getUserFromRequest(r)
			if err != nil {
				log.Println(err)
				unauthorized(w, r)
				return
			}
			if user.ID != image.UserId && !user.isInGroup("admin") && !user.isInGroup("editor") {
				unauthorized(w, r)
				return
			}
		}
	*/
	sendImage(w, r, image.path)
})

var AllUsers = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	users, err := GetAllUsers()
	if err != nil {
		internalError(w, r, err)
		return
	}
	if !user.isInGroup("admin") {
		emptyGroups := make([]string, 0)
		for idx, _ := range users {
			users[idx].Groups = emptyGroups
		}
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"users": users}); err != nil {
		panic(err)
	}
})

var DeleteRow = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	vars := mux.Vars(r)
	rowId, err := strconv.Atoi(vars["rowId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	rowOwnerId, err := GetRowOwnerId(rowId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if rowOwnerId != user.ID && !user.isInGroup("admin") {
		unauthorized(w, r)
		return
	}
	err = RowDelete(rowId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
})

var DeleteUnit = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//this is only called after it is checked that the user is admin, so no checking needed here
	vars := mux.Vars(r)
	unitId, err := strconv.Atoi(vars["unitId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	err = DbDeleteUnit(unitId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
})

var InsertPageResult = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	var pageResult PageResult
	if err := json.Unmarshal(*objmap["pageResult"], &pageResult); err != nil {
		log.Println("not parsable (pageResult)")
		notParsable(w, r, err)
		return
	}
	id, err := DbInsertPageResult(user, pageResult)
	if err != nil {
		internalError(w, r, err)
		return
	}
	pageResult.Id = id
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"pageResult": pageResult}); err != nil {
		panic(err)
	}
})

var UpdatePageResult = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	vars := mux.Vars(r)
	pageResultId, err := strconv.Atoi(vars["pageId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	var pageResult PageResult
	if err := json.Unmarshal(*objmap["pageResult"], &pageResult); err != nil {
		notParsable(w, r, err)
		return
	}
	pageResult.PageId = pageResultId
	err = DbUpdatePageResult(user, pageResult)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("{}")); err != nil {
		panic(err)
	}
})

var GetPageResult = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		log.Println(err)
		unauthorized(w, r)
		return
	}
	vars := mux.Vars(r)
	pageResultId, err := strconv.Atoi(vars["pageResultId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	pageResult, err := DbGetPageResult(pageResultId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if pageResult.UserId != user.ID {
		unauthorized(w, r)
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"pageResult": pageResult}); err != nil {
		panic(err)
	}
})

var InsertUnitResult = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	var unitResult UnitResult
	if err := json.Unmarshal(*objmap["unitResult"], &unitResult); err != nil {
		notParsable(w, r, err)
		return
	}
	err = DbInsertUnitResult(user, unitResult)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"unitResult": unitResult}); err != nil {
		panic(err)
	}
})

var UpdateUnitResult = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	vars := mux.Vars(r)
	unitResultId, err := strconv.Atoi(vars["unitId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	var unitResult UnitResult
	if err := json.Unmarshal(*objmap["unitResult"], &unitResult); err != nil {
		notParsable(w, r, err)
		return
	}
	unitResult.UnitId = unitResultId
	err = DbUpdateUnitResult(user, unitResult)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("{}")); err != nil {
		panic(err)
	}
})

var GetUnitResult = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	unitId, err := strconv.Atoi(vars["unitId"])
	if err != nil {
		notParsable(w, r, err)
		return
	}
	unitResults, err := DbGetUnitResults(unitId)
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"unitResults": unitResults}); err != nil {
		panic(err)
	}
})

var NewPasswordRequest = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var login LoginStruct
	body, err := readBody(r)
	if err != nil {
		internalError(w, r, err)
		return
	}
	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(body, &objmap); err != nil {
		notParsable(w, r, err)
		return
	}
	if err := json.Unmarshal(*objmap["newPasswordRequest"], &login); err != nil {
		notParsable(w, r, err)
		return
	}
	user, err := GetUserByName(login.Username)
	if err != nil {
		log.Println("did not get user")
		log.Println(err)
		w.WriteHeader(http.StatusConflict)
		jsonError := jsonErr{http.StatusConflict, "Username not in DB!"}
		if err := json.NewEncoder(w).Encode(jsonError); err != nil {
			panic(err)
		}
		return
	}
	mailHash, err := HashPWWithSaltB64(login.Email, user.salt)
	b64MailHash := base64.StdEncoding.EncodeToString(mailHash)
	if b64MailHash == user.mailHash {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := make(jwt.MapClaims)

		claims["groups"] = make([]string, 0)
		claims["name"] = user.Username
		claims["exp"] = time.Now().Add(time.Hour * 12).Unix()
		claims["uid"] = user.ID

		token.Claims = claims

		tokenString, _ := token.SignedString(mySigningKey)
		if err := sendMail(login.Email, conf.AppUrl+"password-recovery/"+tokenString, "Objekte im Kreuzverhör: Passwort wiederherstellen", pwRecoveryTemplate); err != nil {
			log.Printf("Error sending mail\n")
			internalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("{}")); err != nil {
			panic(err)
		}
		return
	}
	w.WriteHeader(http.StatusConflict)
	jsonError := jsonErr{http.StatusConflict, "E-Mail and username do not coincide"}
	if err := json.NewEncoder(w).Encode(jsonError); err != nil {
		panic(err)
	}
	return
})
