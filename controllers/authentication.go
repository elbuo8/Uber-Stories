package controllers

import (
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/elbuo8/uber-stories/models"
	"github.com/gorilla/context"
	"github.com/mholt/binding"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"time"
)

const (
	privateKey = ".keys/app.rsa"
)

var (
	signKey []byte
)

func init() {
	var err error
	signKey, err = ioutil.ReadFile(privateKey)
	if err != nil {
		log.Fatal("Error reading Private Key")
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	user := new(models.User)
	bErr := binding.Bind(r, user)
	if bErr != nil {
		BR(w, r, errors.New("Missing information"), http.StatusBadRequest)
		log.Println(bErr)
		return
	}
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r)
		return
	}
	errM := models.CreateUser(dbSession.(*mgo.Session), user)
	if errM != nil {
		if errM.Internal {
			ISR(w, r)
			log.Println(errM.Reason)
			return
		} else {
			BR(w, r, errM.Reason, http.StatusBadRequest)
			return
		}

	}
	context.Set(r, "user", user)
	SetToken(w, r)
}

func LogIn(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	pwd := r.FormValue("password")
	if username == "" || pwd == "" {
		BR(w, r, errors.New("Missing credentials"), http.StatusBadRequest)
		return
	}
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r)
		return
	}
	user, errM := models.AuthUser(dbSession.(*mgo.Session), username, pwd)
	if errM != nil {
		if errM.Internal {
			ISR(w, r)
			log.Println(err) // Make error reporting
			return
		} else {
			BR(w, r, errM.Reason, http.StatusBadRequest)
			return
		}
	}
	context.Set(r, "user", user)
	SetToken(w, r)
}

func SetToken(w http.ResponseWriter, r *http.Request) {
	user, ok := context.GetOk(r, "user")
	if !ok {
		NotAllowed(w, r)
		return
	}
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims["ID"] = user.(*models.User).ID.Hex()
	t.Claims["username"] = user.(*models.User).Username
	t.Claims["exp"] = time.Now().Add(time.Minute * 60 * 730).Unix()
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		ISR(w, r)
		log.Println(err)
		return
	}
	ServeJSON(w, r, &Response{"token": tokenString})
	return
}
