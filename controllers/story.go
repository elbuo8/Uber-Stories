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
	"log"
	"net/http"
)

func StoryHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		vars := mux.Vars(r)
		if user := vars["user"]; user != "" {
			GetUserStories(w, r, user)
		}
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
		ISR(w, r)
		log.Println("Couldn't retrive dbSession from context")
		return
	}
	user := &models.User{
		ID:       bson.ObjectIdHex(tokenInfo["ID"].(string)),
		Username: tokenInfo["username"].(string),
	}
	story, errM := models.NewStory(dbSession.(*mgo.Session), user, content)
	if errM != nil {
		if errM.Internal {
			ISR(w, r)
			log.Println(errM.Reason)
			return
		} else {
			BR(w, r, errM.Reason, http.StatusBadRequest)
		}
	}
	ServeJSON(w, r, &Response{"id": story.ID}, http.StatusCreated)
}

func GetUserStories(w http.ResponseWriter, r *http.Request, username string) {
	dbSession, ok := context.GetOk(r, "dbSession")
	if !ok {
		ISR(w, r)
		log.Println("Couldn't retrive dbSession from context")
		return
	}
	stories, errM := models.StoriesByUser(dbSession.(*mgo.Session), username)
	if errM != nil {
		if errM.Internal {
			ISR(w, r)
			log.Println(errM.Reason)
		} else {
			BR(w, r, errM.Reason, http.StatusBadRequest)
		}
	}
	b, _ := json.Marshal(stories)
	var parse []Response
	json.Unmarshal(b, &parse)
	ServeJSON(w, r, &Response{"stories": parse}, http.StatusOK)
}
