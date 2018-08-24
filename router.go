package main

import (
	"net/http"
	"path"
	"strings"

	"github.com/Mikhalevich/filesharing/handlers"
	"github.com/gorilla/mux"
)

type PublicStorages struct {
	rootPath      string
	permanentPath string
	s             map[string]bool
}

func (p *PublicStorages) Name(r *http.Request) string {
	return mux.Vars(r)["storage"]
}

func (p *PublicStorages) IsPublic(name string) bool {
	_, ok := p.s[name]
	return ok
}

func (p *PublicStorages) Path(name string) string {
	return path.Join(p.rootPath, name)
}

func (p *PublicStorages) PermanentPath(name string) string {
	return path.Join(p.Path(name), p.permanentPath)
}

func NewPublicStorages(root string, permanent string) *PublicStorages {
	return &PublicStorages{
		rootPath:      root,
		permanentPath: permanent,
		s:             map[string]bool{"common": true, "res": true},
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
	h      *handlers.Handlers
}

func NewRouter(p Params) *Router {
	return &Router{
		params: p,
		h:      handlers.NewHandlers(NewPublicStorages(p.RootStorage, p.PermanentDir), params.TempDir),
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
			handler = r.h.CheckAuth(handler)
		}

		if route.StorePath {
			if route.PermanentPath {
				handler = r.h.StorePermanentPath(handler)
			} else {
				handler = r.h.StorePath(handler)
			}
		}

		handler = r.h.RecoverHandler(handler)

		muxRoute.Handler(handler)
	}

	return router
}
