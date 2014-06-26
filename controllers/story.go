package controllers

import (
	"encoding/json"
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/elbuo8/uber-stories/models"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/mail"
	"strconv"
	"time"
)

func StoryHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		GetStories(w, r)
	case r.Method == "PUT":
		SubmitAPI(w, r)
	default:
		http.NotFound(w, r)
	}
}

func SubmitAPI(w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("story")
	if content == "" {
		BR(w, r, errors.New("Missing story"), http.StatusBadRequest)
		return
	}
	token, ok := context.GetOk(r, "token")
	if !ok {
		BR(w, r, errors.New("Missing Token"), http.StatusUnauthorized)
		return
	}
	tokenInfo := token.(*jwt.Token).Claims
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r, errors.New("Couldn't obtain DB Session"))
		return
	}
	user := &models.User{
		ID:       bson.ObjectIdHex(tokenInfo["ID"].(string)),
		Username: tokenInfo["username"].(string),
	}
	story, errM := models.NewStory(dbSession.(*mgo.Session), user, content)
	if errM != nil {
		HandleModelError(w, r, errM)
	}
	ServeJSON(w, r, &Response{"id": story.ID}, http.StatusCreated)
}

func SubmitEmail(w http.ResponseWriter, r *http.Request) {
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r, errors.New("Couldn't obtain DB Session"))
		return
	}
	email, err := mail.ParseAddress(r.FormValue("from"))
	if err != nil {
		ISR(w, r, err)
		return
	}
	if email.Address == "" {
		BR(w, r, errors.New("Missing From Address"), http.StatusBadRequest)
		return
	}
	if r.FormValue("text") == "" {
		BR(w, r, errors.New("Missing Story"), http.StatusBadRequest)
		return
	}
	user, errM := models.FindUserByEmail(dbSession.(*mgo.Session), email.Address)
	if errM != nil {
		HandleModelError(w, r, errM)
	}
	story, errM := models.NewStory(dbSession.(*mgo.Session), user, r.FormValue("text"))
	if errM != nil {
		HandleModelError(w, r, errM)
	}
	ServeJSON(w, r, &Response{"id": story.ID}, http.StatusCreated)
}

func GetStories(w http.ResponseWriter, r *http.Request) {
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r, errors.New("Couldn't obtain DB Session"))
		return
	}
	timeS := r.URL.Query().Get("time")
	var timeV int64
	var err error
	if timeS == "" {
		timeV = time.Now().Unix()
	} else {
		timeV, err = strconv.ParseInt(timeS, 0, 64)
	}
	if err != nil {
		ISR(w, r, err)
		return
	}
	stories, errM := models.GetStories(dbSession.(*mgo.Session), timeV)
	if errM != nil {
		HandleModelError(w, r, errM)
	}
	b, _ := json.Marshal(stories)
	var parse []Response
	json.Unmarshal(b, &parse)
	ServeJSON(w, r, &Response{"stories": parse}, http.StatusOK)
}

func GetUserStories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["user"]
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r, errors.New("Couldn't obtain DB Session"))
		return
	}
	stories, errM := models.StoriesByUser(dbSession.(*mgo.Session), username)
	if errM != nil {
		HandleModelError(w, r, errM)
	}
	b, _ := json.Marshal(stories)
	var parse []Response
	json.Unmarshal(b, &parse)
	ServeJSON(w, r, &Response{"stories": parse}, http.StatusOK)
}
