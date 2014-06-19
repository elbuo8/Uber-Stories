package controllers

import (
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/elbuo8/uber-stories/models"
	"github.com/gorilla/context"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
)

func StoryHandler(w http.ResponseWriter, r *http.Request) {
	switch {
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
	}
	token, ok := context.GetOk(r, "token")
	if !ok {
		ISR(w, r)
		log.Println("Couldn't retrive token from context")
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
	ServeJSON(w, r, &Response{"id": story.ID})
}
