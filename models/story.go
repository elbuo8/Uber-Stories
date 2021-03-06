package models

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type Story struct {
	ID      bson.ObjectId `bson:"_id" json:"-"`
	Author  *User         `bson:"a" json:"owner"`
	Content string        `bson:"cnt" json:"story"`
	Created int64         `bson:"time" json:"createdAt"`
}

func NewStory(s *mgo.Session, a *User, content string) (*Story, *Error) {
	sC := s.DB("uber-stories").C("story")
	story := &Story{ID: bson.NewObjectId(), Author: a, Content: content, Created: time.Now().Unix()}
	err := sC.Insert(story)
	if err != nil {
		return nil, &Error{Reason: err, Internal: true}
	}
	return story, nil
}

func StoriesByUser(s *mgo.Session, username string) ([]*Story, *Error) {
	sC := s.DB("uber-stories").C("story")
	var stories []*Story
	err := sC.Find(bson.M{"a._u": username}).All(&stories)
	if err != nil {
		return nil, &Error{Reason: err, Internal: true}
	}
	return stories, nil
}

func GetStories(s *mgo.Session, t int64) ([]*Story, *Error) {
	sC := s.DB("uber-stories").C("story")
	var stories []*Story
	err := sC.Find(bson.M{"time": bson.M{"$lte": t}}).Sort("Created").Limit(30).All(&stories)

	if err != nil {
		return nil, &Error{Reason: err, Internal: true}
	}
	return stories, nil
}
