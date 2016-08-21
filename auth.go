package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/crypto/scrypt"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

type User struct {
	Username string   `json:"name" db:"username"`
	Groups   []string `json:"groups" db:"groups"`
	Units    []int    `json:"units" db:"units"`
	ID       int      `json:"id" db:"user_id"`
	salt     string
	pwHash   string
	active   bool
}

type LoginStruct struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type RequireRole struct {
	handler http.Handler
	role    string
}

func (u User) isInGroup(group string) bool {
	return stringInSlice(group, u.Groups)
}

func NewRequireRole(handler http.Handler, role string) *RequireRole {
	return &RequireRole{handler, role}
}

func (r *User) UnmarshalJSON(data []byte) error {
	type Alias User
	aux := &struct {
		MyID int `json:"user_id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.ID == 0 {
		r.ID = aux.MyID
	}
	return nil
}

//this only checks if user is in role admin, the jwt has to be checked before this
func (rr *RequireRole) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//check if is in role
	user := context.Get(r, "user")
	claims, ok := user.(*jwt.Token).Claims.(jwt.MapClaims)
	if ok {
		claimGroups, ok := claims["groups"].([]interface{})
		if !ok {
			internalError(w, r, errors.New("could not cast groups(1)"))
			return
		}
		groups := make([]string, len(claimGroups))
		for i, gr := range claimGroups {
			group, ok := gr.(string)
			if !ok {
				internalError(w, r, errors.New("could not cast groups(2)"))
				return
			}
			groups[i] = group
		}
		if stringInSlice(rr.role, groups) {
			rr.handler.ServeHTTP(w, r)
		} else {
			unauthorized(w, r)
			return
		}
	} else {
		internalError(w, r, errors.New("could not read claims"))
	}
}

var mySigningKey = []byte("secret")

var jwtRequiredMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})

var jwtOptionalMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},
	CredentialsOptional: true,
	SigningMethod:       jwt.SigningMethodHS256,
})

func HashPWWithSalt(pw, saltBytes []byte) ([]byte, error) {
	dk, err := scrypt.Key(pw, saltBytes, 16384, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	return dk, nil
}

func HashPWWithSaltB64(pw, salt string) ([]byte, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, err
	}
	return HashPWWithSalt([]byte(pw), saltBytes)
}

func HashNewPW(pw string) (salt string, hash []byte, err error) {
	saltBytes := make([]byte, 16)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return "", nil, err
	}
	dk, err := HashPWWithSalt([]byte(pw), saltBytes)
	return base64.StdEncoding.EncodeToString(saltBytes), dk, err
}
