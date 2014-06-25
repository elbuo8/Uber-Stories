package main

import (
	"./controllers"
	"./models"
	"errors"
	"github.com/codegangsta/negroni"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	logrus "github.com/meatballhat/negroni-logrus"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

const (
	publicKey = ".keys/app.rsa.pub"
)

var (
	verifyKey []byte
)

func init() {
	var err error
	verifyKey, err = ioutil.ReadFile(publicKey)
	if err != nil {
		log.Fatal("Error reading Private key")
		return
	}
}

func BuildApp() *negroni.Negroni {
	r := mux.NewRouter()
	n := negroni.Classic()
	dbSession := InitDB()
	SetGandalf(n)
	SetDB(dbSession, n)
	SetMiddleware(n)
	SetRoutes(r)
	n.UseHandler(r)
	return n
}

func InitDB() *mgo.Session {
	db, err := mgo.Dial(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Println("%v captured - Closing database connection", sig)
			db.Close()
			os.Exit(1)
		}
	}()
	err = models.Setup(db)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func SetGandalf(n *negroni.Negroni) {
	n.Use(negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if h := r.Header.Get("Authorization"); h != "" {
			token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) ([]byte, error) {
				return verifyKey, nil
			})
			// https://gist.github.com/cryptix/45c33ecf0ae54828e63b
			switch err.(type) {
			case nil:
				if !token.Valid {
					controllers.NotAllowed(w, r)
					return
				}
				context.Set(r, "token", token)
				next(w, r)
			case *jwt.ValidationError:
				vErr := err.(*jwt.ValidationError)
				switch vErr.Errors {
				case jwt.ValidationErrorExpired:
					controllers.BR(w, r, errors.New("Token Expired"), http.StatusUnauthorized)
					return
				default:
					controllers.ISR(w, r, errors.New(vErr.Error()))
					log.Println(vErr.Error())
					return
				}
			default:
				controllers.ISR(w, r, err)
				return
			}
		} else {
			next(w, r)
		}
	}))
}

func SetDB(s *mgo.Session, n *negroni.Negroni) {
	n.Use(negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		context.Set(r, "dbSession", s.Clone())
		next(w, r)
	}))
}

func SetMiddleware(n *negroni.Negroni) {
	n.Use(negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		err := r.ParseForm()
		if err != nil {
			controllers.ISR(w, r, err)
			return
		}
		next(w, r)
		if strings.Contains(r.Header.Get("Content-Type"), "multipart") {
			err = r.ParseMultipartForm(1024)
			if err != nil {
				controllers.ISR(w, r, err)
				return
			}
		}
	}))
	n.Use(logrus.NewMiddleware())
}

func SetRoutes(r *mux.Router) {
	r.HandleFunc("/auth", controllers.AuthorizationHandler).Methods("PUT POST")
	r.HandleFunc("/verify/{id}", controllers.VerifyEmail).Methods("GET")
	r.HandleFunc("/api/story", controllers.StoryHandler).Methods("GET PUT")
	r.HandleFunc("/api/story/{user}", controllers.GetUserStories).Methods("GET")
	r.HandleFunc("/api/user", controllers.UserHandler).Methods("GET")
	r.HandleFunc("/webhook/email", controllers.SubmitEmail).Methods("POST")
}

func main() {
	log.Fatal(http.ListenAndServe(":3000", BuildApp()))
}
