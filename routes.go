package main

import (
	"fmt"
	"log"
	"net/http"
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

var (
	staticStorages map[string]bool = map[string]bool{"common": true, "res": true}
	routes                         = Routes{
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
			"/register/",
			false,
			"GET,POST",
			false,
			http.HandlerFunc(registerHandler),
		},
		Route{
			"/login/{storage}/",
			false,
			"GET,POST",
			false,
			http.HandlerFunc(loginHandler),
		},
		Route{
			"/{storage}/permanent/",
			false,
			"GET",
			true,
			http.HandlerFunc(viewPermanentHandler),
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
			http.HandlerFunc(uploadStorageHandler),
		},
		Route{
			"/{storage}/permanent/upload/",
			false,
			"POST",
			true,
			http.HandlerFunc(uploadPermanentHandler),
		},
		Route{
			"/{storage}/remove/",
			false,
			"POST",
			true,
			http.HandlerFunc(removeStorageHandler),
		},
		Route{
			"/{storage}/permanent/remove/",
			false,
			"POST",
			true,
			http.HandlerFunc(removePermanentHandler),
		},
		Route{
			"/{storage}/shareText/",
			false,
			"POST",
			true,
			http.HandlerFunc(shareTextStorageHandler),
		},
		Route{
			"/{storage}/permanent/shareText/",
			false,
			"POST",
			true,
			http.HandlerFunc(shareTextPermanentHandler),
		},
	}
)

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
		storageName := storageVar(r)
		var err error

		if !needAuth {
			next.ServeHTTP(w, r)
			return
		}

		if storageName == "" {
			log.Println(fmt.Sprintf("Storage name is empty for %s", r.URL))
			next.ServeHTTP(w, r)
			return
		}

		if _, ok := staticStorages[storageName]; ok {
			err = checkStorage(storageName, false)
			if err != nil {
				respondError(err, w, http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		storage := NewStorage()
		defer storage.Close()

		user, err := storage.UserByName(storageName)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		err = checkStorage(storageName, true)
		if err != nil {
			respondError(err, w, http.StatusInternalServerError)
			return
		}

		if user.Password.isEmpty() {
			next.ServeHTTP(w, r)
			return
		}

		authorized := false
		defer func() {
			if authorized {
				next.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, "/login/"+storageName, http.StatusFound)
			}
		}()

		cookies := r.Cookies()
		for _, cook := range cookies {
			if cook.Name == storageName {
				session, err := user.SessionById(cook.Value)
				if err != nil {
					removeCookie(w, storageName)
					return
				}

				if isExpired(session.Expires) {
					removeCookie(w, storageName)
					return
				}

				authorized = true
				return
			}
		}
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
