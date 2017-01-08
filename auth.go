package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"text/template"

	"golang.org/x/crypto/scrypt"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

const mailTemplate = `To: {{.Recipient}}
Subject: Registrierung Objekte im Kreuzverhör

Bitte klicken Sie auf den Folgenden Link um Ihre Registrierung abzuschließen:
{{.TokenLink}}`

type User struct {
	Username string   `json:"name" db:"username"`
	Groups   []string `json:"groups" db:"groups"`
	Units    []int    `json:"units" db:"units"`
	ID       int      `json:"id" db:"user_id"`
	salt     string
	pwHash   string
	Active   bool `json:"active"`
	mailHash string
	Points   uint `json:"points"`
	Rank     uint `json:"rank"`
}

type LoginStruct struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Email    string `json:"email" db:"email"`
}

type MailTemplate struct {
	Recipient string
	TokenLink string
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

func getUserFromRequest(r *http.Request) (User, error) {
	userJWT := context.Get(r, "user")
	if userJWT == nil {
		return User{}, errors.New("no token in context")
	}
	claims, ok := userJWT.(*jwt.Token).Claims.(jwt.MapClaims)
	if ok {
		claimIdF, ok := claims["uid"].(float64)
		if !ok {
			return User{}, errors.New("could not cast uid to int")
		}
		claimId := int(claimIdF)
		claimGroups, ok := claims["groups"].([]interface{})
		if !ok {
			return User{}, errors.New("could not parse groups")
		}
		groups := make([]string, len(claimGroups))
		for i, gr := range claimGroups {
			group, ok := gr.(string)
			if !ok {
				return User{}, errors.New("could not parse groups")
			}
			groups[i] = group
		}
		name, ok := claims["name"].(string)
		if !ok {
			return User{}, errors.New("could not parse name")
		}
		return User{Username: name, Groups: groups, ID: claimId}, nil

	} else {
		return User{}, errors.New("could not read claims")
	}
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

func GetJWTClaims(r *http.Request) (jwt.MapClaims, error) {
	user := context.Get(r, "user")
	claims, ok := user.(*jwt.Token).Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not cast claims")
	} else {
		return claims, nil
	}
}

func GetClaimGroups(claims jwt.MapClaims) ([]string, error) {
	claimGroups, ok := claims["groups"].([]interface{})
	if !ok {
		return nil, errors.New("could not cas groups(1)")
	}
	groups := make([]string, len(claimGroups))
	for i, gr := range claimGroups {
		group, ok := gr.(string)
		if !ok {
			return nil, errors.New("could not cast groups(2)")
		}
		groups[i] = group
	}
	return groups, nil
}

//this only checks if user is in given role, the jwt has to be checked before this
func (rr *RequireRole) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//check if is in role
	user := context.Get(r, "user")
	claims, ok := user.(*jwt.Token).Claims.(jwt.MapClaims)
	if ok {
		groups, err := GetClaimGroups(claims)
		if err != nil {
			internalError(w, r, err)
			return
		}
		if stringInSlice(rr.role, groups) {
			rr.handler.ServeHTTP(w, r)
		} else {
			log.Println("user not in role")
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

func sendRegistrationMail(recipient string, token string) error {
	auth := smtp.PlainAuth("", conf.MailConfig.UserName, conf.MailConfig.Password, conf.MailConfig.Host)
	to := []string{recipient}
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	mailInfo := MailTemplate{recipient, token}
	var doc bytes.Buffer
	t, err := template.New("mail").Parse(mailTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(&doc, mailInfo)
	if err != nil {
		return err
	}
	err = smtp.SendMail(conf.MailConfig.Host+":"+strconv.Itoa(conf.MailConfig.Port), auth, conf.MailConfig.From, to, doc.Bytes())
	if err != nil {
		return err
	}
	return nil
}
