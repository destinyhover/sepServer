package server

import (
	"net/http"
	_ "restapi/cmd/internal/data"
	"time"

	"github.com/gorilla/mux"
)

var rMux = mux.NewRouter()
var Port = ":1234"

func GetServer() *http.Server {
	s := http.Server{
		Addr:         Port,
		Handler:      rMux,
		ErrorLog:     nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	rMux.NotFoundHandler = http.HandlerFunc(DefaultHandler)

	notAllowed := notAllowedHandler{}

	rMux.MethodNotAllowedHandler = notAllowed

	rMux.HandleFunc("/time", TimeHandler)

	// 	Register GET methods
	getMux := rMux.Methods(http.MethodGet).Subrouter()

	getMux.HandleFunc("/getall", GetAllHandler)
	getMux.HandleFunc("/getid/{username}", GetIdHandler)
	getMux.HandleFunc("/logged", GetLoggedHandler)
	getMux.HandleFunc("/username/{id:[0-9]+}", GetUserDataHandler)

	// Register PUT
	purMux := rMux.Methods(http.MethodPut).Subrouter()
	purMux.HandleFunc("/update", UpdateHandler)

	// Register POST
	postMux := rMux.Methods(http.MethodPost).Subrouter()
	postMux.HandleFunc("/add", AddHandler)
	postMux.HandleFunc("/login", LoginHandler)
	postMux.HandleFunc("/logout", LogoutHandler)

	// Register delete
	deleteMux := rMux.Methods(http.MethodDelete).Subrouter()
	deleteMux.HandleFunc("/username/{id:[0-9]+}", DeleteHandler)
	return &s
}
