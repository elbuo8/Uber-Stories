package controllers

import (
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/elbuo8/uber-stories/models"
	"github.com/elbuo8/uber-stories/services"
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
		ISR(w, r, errors.New("Couldn't obtain DB Session"))
		return
	}
	errM := models.CreateUser(dbSession.(*mgo.Session), user)
	if errM != nil {
		HandleModelError(w, r, errM)
		return
	}
	go services.ActivationEmail(user)
	SetToken(w, r, user)
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r, errors.New("Couldn't obtain DB Session"))
		return
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		BR(w, r, errors.New("Missing Hash"), http.StatusBadRequest)
		return
	}
	errM := models.VerifyEmail(dbSession.(*mgo.Session), bson.ObjectIdHex(id))
	if errM != nil {
		HandleModelError(w, r, errM)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
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
		ISR(w, r, errors.New("Couldn't obtain DB Session"))
		return
	}
	user, errM := models.AuthUser(dbSession.(*mgo.Session), username, pwd)
	if errM != nil {
		HandleModelError(w, r, errM)
		return
	}
	SetToken(w, r, user)
}

func SetToken(w http.ResponseWriter, r *http.Request, user *models.User) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims["ID"] = user.ID.Hex()
	t.Claims["username"] = user.Username
	t.Claims["exp"] = time.Now().Add(time.Minute * 60 * 730).Unix()
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		ISR(w, r, err)
		return
	}
	ServeJSON(w, r, &Response{"token": tokenString}, http.StatusOK)
	return
}
