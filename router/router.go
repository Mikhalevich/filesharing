package router

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
	Pattern            string
	IsPrefix           bool
	Methods            string
	NeedAuth           bool
	StorePath          bool
	StorePermanentPath bool
	Handler            http.Handler
}

type Router struct {
	rootStorage string
	enableAuth  bool
	routes      []Route
	h           *handlers.Handlers
}

func NewRouter(root string, ea bool, handl *handlers.Handlers) *Router {
	return &Router{
		rootStorage: root,
		enableAuth:  ea,
		h:           handl,
	}
}

func (r *Router) makeRoutes() {
	r.routes = []Route{
		Route{
			Pattern: "/",
			Methods: "GET",
			Handler: http.HandlerFunc(r.h.RootHandler),
		},
		Route{
			Pattern:  "/res/",
			IsPrefix: true,
			Methods:  "GET",
			Handler:  http.StripPrefix("/res/", http.FileServer(http.Dir("res"))),
		},
		Route{
			Pattern: "/register/",
			Methods: "GET,POST",
			Handler: http.HandlerFunc(r.h.RegisterHandler),
		},
		Route{
			Pattern:   "/login/{storage}/",
			Methods:   "GET,POST",
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.LoginHandler),
		},
		Route{
			Pattern:            "/api/{storage}/permanent/",
			Methods:            "GET",
			StorePermanentPath: true,
			Handler:            http.HandlerFunc(r.h.JSONViewHandler),
		},
		Route{
			Pattern:   "/api/{storage}/",
			Methods:   "GET",
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.JSONViewHandler),
		},
		Route{
			Pattern:   "/{storage}/index.html",
			Methods:   "GET",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.IndexHTMLHandler),
		},
		Route{
			Pattern:            "/{storage}/permanent/index.html",
			Methods:            "GET",
			NeedAuth:           true,
			StorePermanentPath: true,
			Handler:            http.HandlerFunc(r.h.IndexHTMLHandler),
		},
		Route{
			Pattern:            "/{storage}/permanent/",
			Methods:            "GET",
			NeedAuth:           true,
			StorePermanentPath: true,
			Handler:            http.HandlerFunc(r.h.ViewHandler),
		},
		Route{
			Pattern:   "/{storage}/",
			Methods:   "GET",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.ViewHandler),
		},
		Route{
			Pattern:  "/{storage}/",
			IsPrefix: true,
			Methods:  "GET",
			NeedAuth: true,
			Handler:  http.FileServer(http.Dir(r.rootStorage)),
		},
		Route{
			Pattern:   "/{storage}/upload/",
			Methods:   "POST",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.UploadHandler),
		},
		Route{
			Pattern:            "/{storage}/permanent/upload/",
			Methods:            "POST",
			NeedAuth:           true,
			StorePermanentPath: true,
			Handler:            http.HandlerFunc(r.h.UploadHandler),
		},
		Route{
			Pattern:   "/{storage}/remove/",
			Methods:   "POST",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.RemoveHandler),
		},
		Route{
			Pattern:            "/{storage}/permanent/remove/",
			Methods:            "POST",
			NeedAuth:           true,
			StorePermanentPath: true,
			Handler:            http.HandlerFunc(r.h.RemoveHandler),
		},
		Route{
			Pattern:   "/{storage}/shareText/",
			Methods:   "POST",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.ShareTextHandler),
		},
		Route{
			Pattern:            "/{storage}/permanent/shareText/",
			Methods:            "POST",
			NeedAuth:           true,
			StorePermanentPath: true,
			Handler:            http.HandlerFunc(r.h.ShareTextHandler),
		},
	}
}

func (r *Router) Handler() http.Handler {
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
		if r.enableAuth && route.NeedAuth {
			handler = r.h.CheckAuth(handler)
		}

		if route.StorePath || route.StorePermanentPath {
			if route.StorePermanentPath {
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
