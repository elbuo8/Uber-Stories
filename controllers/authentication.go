package controllers

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/elbuo8/uber-stories/models"
	"github.com/gorilla/context"
	"github.com/mholt/binding"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"time"
)

var (
	signKey []byte
)

func init() {
	signKey = []byte("Hello")
}

func Register(w http.ResponseWriter, r *http.Request) {
	user := new(models.User)
	bErr := binding.Bind(r, user)
	if bErr != nil {
		BR(w, r)
		log.Println(bErr)
		return
	}
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r)
		return
	}
	err := models.CreateUser(dbSession.(*mgo.Session), user)
	if err != nil {
		// Make error reporting
		log.Println(err)
		ISR(w, r)
		return
	}
	context.Set(r, "user", user)
	SetToken(w, r)
}

func LogIn(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		ISR(w, r)
		log.Println(err)
		return
	}
	username := r.FormValue("username")
	pwd := r.FormValue("password")
	if username == "" || pwd == "" {
		BR(w, r)
		return
	}
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r)
		return
	}
	user, err := models.AuthUser(dbSession.(*mgo.Session), username, pwd)
	if err != nil {
		BR(w, r)
		log.Println(err) // Make error reporting
		return
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
	t := jwt.New(jwt.GetSigningMethod("HS256"))
	t.Claims["ID"] = user.(*models.User).ID
	t.Claims["exp"] = time.Now().Add(time.Minute * 60 * 730).Unix()
	log.Println(signKey)
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		ISR(w, r)
		log.Println(err)
		return
	}
	ServeJSON(w, r, &Response{"token": tokenString})
	return
}
