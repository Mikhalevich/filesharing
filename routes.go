package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Pattern     string
	Methods     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"/",
		"GET",
		rootHandler,
	},
	Route{
		"/upload/",
		"POST",
		uploadHandler,
	},
	Route{
		"/remove/",
		"POST",
		removeHandler,
	},
	Route{
		"/shareText/",
		"POST",
		shareTextHandler,
	},
}

func recoverHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)

				return
			}
		}()

		next.ServeHTTP(w, r)
	}
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Path(route.Pattern).
			Methods(strings.Split(route.Methods, ",")...).
			Handler(recoverHandler(route.HandlerFunc))
	}

	// static resourses
	router.PathPrefix("/res/").Handler(http.StripPrefix("/res/", http.FileServer(http.Dir("res"))))
	router.PathPrefix("/storage/").Handler(http.StripPrefix("/storage/", http.FileServer(http.Dir(storageDir))))

	return router
}
