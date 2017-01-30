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
		"/upload",
		"POST",
		uploadHandler,
	},
	Route{
		"/remove",
		"POST",
		removeHandler,
	},
	Route{
		"/shareText",
		"POST",
		shareTextHandler,
	},
}

func recoverHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)

				return
			}
		}()

		fn(w, r)
	}
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(strings.Split(route.Methods, ",")...).
			Path(route.Pattern).
			Handler(recoverHandler(route.HandlerFunc))
	}

	return router
}
