package models

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type Story struct {
	ID      bson.ObjectId `bson:"_id"`
	Author  *User         `bson:"a"`
	Content string        `bson:"cnt"`
	Created time.Time     `bson:"time"`
}

func NewStory(s *mgo.Session, a *User, content string) (*Story, *Error) {
	sC := s.DB("uber-stories").C("story")
	story := &Story{ID: bson.NewObjectId(), Author: a, Content: content, Created: time.Now()}
	err := sC.Insert(story)
	if err != nil {
		return nil, &Error{Reason: err, Internal: true}
	}
	return story, nil
}
