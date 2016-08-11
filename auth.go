package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/crypto/scrypt"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Username string   `json:"username" db:"username"`
	Groups   []string `json:"groups" db:"groups"`
	Salt     string   `json:"salt" db:"salt"`
	PWHash   string   `json:"pwhash" db:"pwhash"`
	Active   bool     `json:"active" db:"active"`
	ID       int      `json:"id" db:"user_id"`
}

type LoginStruct struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
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

var mySigningKey = []byte("secret")

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},
	SigningMethod: jwt.SigningMethodHS256,
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
