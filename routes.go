package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Mikhalevich/filesharing/db"
	"github.com/gorilla/context"
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

const (
	contextStoragePath = "storagePath"
)

var (
	staticStorages map[string]bool = map[string]bool{"common": true, "res": true}
	routes         Routes
)

func makeRoutes() {
	routes = Routes{
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
			storePath(http.HandlerFunc(loginHandler)),
		},
		Route{
			"/{storage}/index.html",
			false,
			"GET",
			true,
			storePath(http.HandlerFunc(indexHTMLHandler)),
		},
		Route{
			"/{storage}/permanent/index.html",
			false,
			"GET",
			true,
			storePermanentPath(http.HandlerFunc(indexHTMLHandler)),
		},
		Route{
			"/{storage}/permanent/",
			false,
			"GET",
			true,
			storePermanentPath(http.HandlerFunc(viewHandler)),
		},
		Route{
			"/{storage}/",
			false,
			"GET",
			true,
			storePath(http.HandlerFunc(viewHandler)),
		},
		Route{
			"/{storage}/",
			true,
			"GET",
			true,
			http.FileServer(http.Dir(params.RootStorage)),
		},
		Route{
			"/{storage}/upload/",
			false,
			"POST",
			true,
			storePath(http.HandlerFunc(uploadHandler)),
		},
		Route{
			"/{storage}/permanent/upload/",
			false,
			"POST",
			true,
			storePermanentPath(http.HandlerFunc(uploadHandler)),
		},
		Route{
			"/{storage}/remove/",
			false,
			"POST",
			true,
			storePath(http.HandlerFunc(removeHandler)),
		},
		Route{
			"/{storage}/permanent/remove/",
			false,
			"POST",
			true,
			storePermanentPath(http.HandlerFunc(removeHandler)),
		},
		Route{
			"/{storage}/shareText/",
			false,
			"POST",
			true,
			storePath(http.HandlerFunc(shareTextHandler)),
		},
		Route{
			"/{storage}/permanent/shareText/",
			false,
			"POST",
			true,
			storePermanentPath(http.HandlerFunc(shareTextHandler)),
		},
	}
}

func contextStorage(r *http.Request) string {
	return context.Get(r, contextStoragePath).(string)
}

func storePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storage := mux.Vars(r)["storage"]
		if storage == "" {
			log.Printf("Invalid storage request, url = %s", r.URL)
		} else {
			context.Set(r, contextStoragePath, storagePath(storage))
		}

		next.ServeHTTP(w, r)
	})
}

func storePermanentPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storage := mux.Vars(r)["storage"]
		if storage == "" {
			log.Printf("Invalid storage request, url = %s", r.URL)
		} else {
			context.Set(r, contextStoragePath, permanentPath(storage))
		}

		next.ServeHTTP(w, r)
	})
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

func noAuth(next http.Handler, needAuth bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		storage := db.NewStorage()
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

		if user.Password.IsEmpty() {
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

func NewRouter(allowPrivate bool) *mux.Router {
	makeRoutes()

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		muxRoute := router.NewRoute()
		if route.IsPrefix {
			muxRoute.PathPrefix(route.Pattern)
		} else {
			muxRoute.Path(route.Pattern)
		}
		authFunc := noAuth
		if allowPrivate {
			authFunc = checkAuth
		}
		muxRoute.Methods(strings.Split(route.Methods, ",")...)
		muxRoute.Handler(recoverHandler(authFunc(route.Handler, route.NeedAuth)))
	}

	return router
}
