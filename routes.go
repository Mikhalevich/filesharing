package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Pattern  string
	IsPrefix bool
	Methods  string
	NeedAuth bool
	Handler  http.Handler
}

type Routes []Route

var routes = Routes{
	Route{
		"/",
		false,
		"GET",
		false,
		http.HandlerFunc(rootHandler),
	},
	Route{
		"/res/",
		true,
		"GET",
		false,
		http.StripPrefix("/res/", http.FileServer(http.Dir("res"))),
	},
	Route{
		"/{storage}/",
		false,
		"GET",
		true,
		http.HandlerFunc(viewStorageHandler),
	},
	Route{
		"/{storage}/",
		true,
		"GET",
		true,
		http.FileServer(http.Dir(rootStorageDir)),
	},
	Route{
		"/{storage}/upload/",
		false,
		"POST",
		true,
		http.HandlerFunc(uploadHandler),
	},
	Route{
		"/{storage}/remove/",
		false,
		"POST",
		true,
		http.HandlerFunc(removeHandler),
	},
	Route{
		"/{storage}/shareText/",
		false,
		"POST",
		true,
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

func checkAuth(next http.Handler, needAuth bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !needAuth {
			next.ServeHTTP(w, r)
		}

		storageName := storageVar(r)
		_, err := os.Stat(storagePath(storageName))
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
			} else {
				respondError(err, w, http.StatusInternalServerError)
			}
			return
		}

		//todo: check auth

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
		muxRoute.Handler(recoverHandler(checkAuth(route.Handler, route.NeedAuth)))
	}

	return router
}
