package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type StorageChecker interface {
	IsPublic(name string) bool
}

type PublicStorages struct {
	s map[string]bool
}

func (p *PublicStorages) IsPublic(name string) bool {
	_, ok := p.s[name]
	return ok
}

func NewPublicStorages() *PublicStorages {
	return &PublicStorages{
		s: map[string]bool{"common": true, "res": true},
	}
}

type Route struct {
	Pattern       string
	IsPrefix      bool
	Methods       string
	NeedAuth      bool
	StorePath     bool
	PermanentPath bool
	Handler       http.Handler
}

type Router struct {
	params Params
	routes []Route
	h      *Handlers
}

func NewRouter(p Params) *Router {
	return &Router{
		params: p,
		h:      NewHandlers(NewPublicStorages()),
	}
}

func (r *Router) makeRoutes() {
	r.routes = []Route{
		Route{
			"/",
			false,
			"GET",
			false,
			false,
			false,
			http.HandlerFunc(r.h.RootHandler),
		},
		Route{
			"/res/",
			true,
			"GET",
			false,
			false,
			false,
			http.StripPrefix("/res/", http.FileServer(http.Dir("res"))),
		},
		Route{
			"/register/",
			false,
			"GET,POST",
			false,
			false,
			false,
			http.HandlerFunc(r.h.RegisterHandler),
		},
		Route{
			"/login/{storage}/",
			false,
			"GET,POST",
			false,
			true,
			false,
			http.HandlerFunc(r.h.LoginHandler),
		},
		Route{
			"/{storage}/index.html",
			false,
			"GET",
			true,
			true,
			false,
			http.HandlerFunc(r.h.IndexHTMLHandler),
		},
		Route{
			"/{storage}/permanent/index.html",
			false,
			"GET",
			true,
			true,
			true,
			http.HandlerFunc(r.h.IndexHTMLHandler),
		},
		Route{
			"/{storage}/permanent/",
			false,
			"GET",
			true,
			true,
			true,
			http.HandlerFunc(r.h.ViewHandler),
		},
		Route{
			"/{storage}/",
			false,
			"GET",
			true,
			true,
			false,
			http.HandlerFunc(r.h.ViewHandler),
		},
		Route{
			"/{storage}/",
			true,
			"GET",
			true,
			false,
			false,
			http.FileServer(http.Dir(params.RootStorage)),
		},
		Route{
			"/{storage}/upload/",
			false,
			"POST",
			true,
			true,
			false,
			http.HandlerFunc(r.h.UploadHandler),
		},
		Route{
			"/{storage}/permanent/upload/",
			false,
			"POST",
			true,
			true,
			true,
			http.HandlerFunc(r.h.UploadHandler),
		},
		Route{
			"/{storage}/remove/",
			false,
			"POST",
			true,
			true,
			false,
			http.HandlerFunc(r.h.RemoveHandler),
		},
		Route{
			"/{storage}/permanent/remove/",
			false,
			"POST",
			true,
			true,
			true,
			http.HandlerFunc(r.h.RemoveHandler),
		},
		Route{
			"/{storage}/shareText/",
			false,
			"POST",
			true,
			true,
			false,
			http.HandlerFunc(r.h.ShareTextHandler),
		},
		Route{
			"/{storage}/permanent/shareText/",
			false,
			"POST",
			true,
			true,
			true,
			http.HandlerFunc(r.h.ShareTextHandler),
		},
	}
}

func (r *Router) handler() http.Handler {
	r.makeRoutes()

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range r.routes {
		muxRoute := router.NewRoute()
		if route.IsPrefix {
			muxRoute.PathPrefix(route.Pattern)
		} else {
			muxRoute.Path(route.Pattern)
		}

		muxRoute.Methods(strings.Split(route.Methods, ",")...)

		handler := route.Handler
		if r.params.AllowPrivate && route.NeedAuth {
			handler = r.h.checkAuth(handler)
		}

		if route.StorePath {
			if route.PermanentPath {
				handler = r.h.storePermanentPath(handler)
			} else {
				handler = r.h.storePath(handler)
			}
		}

		handler = r.h.recoverHandler(handler)

		muxRoute.Handler(handler)
	}

	return router
}
