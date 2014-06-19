package main

import (
	"./controllers"
	"./models"
	"github.com/codegangsta/negroni"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	logrus "github.com/meatballhat/negroni-logrus"
	"labix.org/v2/mgo"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

var (
	verifyKey []byte
)

func init() {
	// Make this proper later
	verifyKey = []byte("bye")
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
		if h := r.Header.Get("Authorization"); strings.HasPrefix(r.URL.Path, "/api") && h != "" { // Token Required
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
					controllers.ISR(w, r)
					return
				default:
					controllers.ISR(w, r)
					return
				}
			default:
				controllers.ISR(w, r)
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
	n.Use(logrus.NewMiddleware())
}

func SetRoutes(r *mux.Router) {
	r.HandleFunc("/register", controllers.Register)
	r.HandleFunc("/login", controllers.LogIn)
}

func main() {
	log.Fatal(http.ListenAndServe(":3000", BuildApp()))
}
