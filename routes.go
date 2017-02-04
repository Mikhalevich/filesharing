package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Pattern  string
	IsPrefix bool
	Methods  string
	Handler  http.Handler
}

type Routes []Route

var routes = Routes{
	Route{
		"/",
		false,
		"GET",
		http.HandlerFunc(rootHandler),
	},
	Route{
		"/res/",
		true,
		"GET",
		http.StripPrefix("/res/", http.FileServer(http.Dir("res"))),
	},
	Route{
		"/{storage}/",
		false,
		"GET",
		http.HandlerFunc(viewStorageHandler),
	},
	Route{
		"/{storage}/",
		true,
		"GET",
		http.FileServer(http.Dir(rootStorageDir)),
	},
	Route{
		"/{storage}/upload/",
		false,
		"POST",
		http.HandlerFunc(uploadHandler),
	},
	Route{
		"/{storage}/remove/",
		false,
		"POST",
		http.HandlerFunc(removeHandler),
	},
	Route{
		"/{storage}/shareText/",
		false,
		"POST",
		http.HandlerFunc(shareTextHandler),
	},
}

func recoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)

				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		muxRoute := router.NewRoute()
		if route.IsPrefix {
			muxRoute.PathPrefix(route.Pattern)
		} else {
			muxRoute.Path(route.Pattern)
		}
		muxRoute.Methods(strings.Split(route.Methods, ",")...)
		muxRoute.Handler(recoverHandler(route.Handler))
	}

	return router
}
