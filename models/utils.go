package models

import (
	"labix.org/v2/mgo"
)

type Error struct {
	Reason   error
	Internal bool
}

func Setup(s *mgo.Session) error {
	i := mgo.Index{
		Key:        []string{"u"},
		Unique:     true,
		Background: true,
		Name:       "u",
	}
	return s.DB("uber-stories").C("user").EnsureIndex(i)
}
