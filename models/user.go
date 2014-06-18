package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	"errors"
	"github.com/mholt/binding"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type User struct {
	ID       bson.ObjectId `bson:"_id"`
	Username string        `bson:"u"`
	Password string        `bson:"pwd"`
	Email    string        `bson:"mail"`
}

func (u *User) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&u.Email: binding.Field{
			Form:     "email",
			Required: true,
		},
		&u.Username: binding.Field{
			Form:     "username",
			Required: true,
		},
		&u.Password: binding.Field{
			Form:     "password",
			Required: true,
		},
	}
}

func CreateUser(s *mgo.Session, u *User) error {
	uC := s.DB("uber-stories").C("user")
	pwHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("500 - Error hashing password")
	}
	u.Password = string(pwHash)
	u.ID = bson.NewObjectId()
	err = uC.Insert(u)
	if mgo.IsDup(err) {
		return errors.New("302 - Existing user error")
	}
	return nil
}

func AuthUser(s *mgo.Session, u, pwd string) (*User, error) {
	uC := s.DB("uber-stories").C("user")
	user := &User{}
	err := uC.Find(bson.M{"u": u}).One(user)
	if err != nil {
		return nil, err
	}
	if user.ID == "" {
		return nil, nil
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pwd))
	if err != nil {
		return nil, errors.New("Incorrect password")
	}
	return user, nil
}
