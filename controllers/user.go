package controllers

import (
	"encoding/json"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/elbuo8/uber-stories/models"
	"github.com/gorilla/context"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		GetUser(w, r)
	default:
		http.NotFound(w, r)
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
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
	log.Println(bson.ObjectIdHex(tokenInfo["ID"].(string)))
	user, errM := models.FindUser(dbSession.(*mgo.Session), bson.ObjectIdHex(tokenInfo["ID"].(string)))
	if errM != nil {
		if errM.Internal {
			ISR(w, r)
			log.Println(errM.Reason)
		} else {
			BR(w, r, errM.Reason, http.StatusBadRequest)
		}
	}
	// Stupid hack. Fix.
	b, _ := json.Marshal(user)
	parse := &Response{}
	json.Unmarshal(b, parse)
	log.Println(parse)
	ServeJSON(w, r, parse, http.StatusOK)
}
