package router

import (
	"log"
	"net/http"
	"strings"

	"github.com/Mikhalevich/filesharing/handlers"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

const (
	ContextStorageName        = "storageName"
	ContextStorageIsPermanent = "storageIsPermanent"
)

type PublicStorages struct {
	rootPath      string
	permanentPath string
	s             map[string]bool
}

func (p *PublicStorages) Name(r *http.Request) string {
	return mux.Vars(r)["storage"]
}

func (p *PublicStorages) IsPermanent(r *http.Request) bool {
	_, ok := context.GetOk(r, ContextStorageIsPermanent)
	return ok
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
	enableAuth bool
	routes     []Route
	h          *handlers.Handlers
}

func NewRouter(ea bool, handl *handlers.Handlers) *Router {
	return &Router{
		enableAuth: ea,
		h:          handl,
	}
}

func (r *Router) makeRoutes() {
	r.routes = []Route{
		Route{
			Pattern: "/",
			Methods: "GET",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/common/", http.StatusMovedPermanently)
			}),
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
			Pattern:       "/api/{storage}/permanent/",
			Methods:       "GET",
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.JSONViewHandler),
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
			Pattern:       "/{storage}/permanent/index.html",
			Methods:       "GET",
			NeedAuth:      true,
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.IndexHTMLHandler),
		},
		Route{
			Pattern:       "/{storage}/permanent/",
			Methods:       "GET",
			NeedAuth:      true,
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.ViewHandler),
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
			Handler:  r.h.FileServer(),
		},
		Route{
			Pattern:   "/{storage}/upload/",
			Methods:   "POST",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.UploadHandler),
		},
		Route{
			Pattern:       "/{storage}/permanent/upload/",
			Methods:       "POST",
			NeedAuth:      true,
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.UploadHandler),
		},
		Route{
			Pattern:   "/{storage}/remove/",
			Methods:   "POST",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.RemoveHandler),
		},
		Route{
			Pattern:       "/{storage}/permanent/remove/",
			Methods:       "POST",
			NeedAuth:      true,
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.RemoveHandler),
		},
		Route{
			Pattern:   "/{storage}/shareText/",
			Methods:   "POST",
			NeedAuth:  true,
			StorePath: true,
			Handler:   http.HandlerFunc(r.h.ShareTextHandler),
		},
		Route{
			Pattern:       "/{storage}/permanent/shareText/",
			Methods:       "POST",
			NeedAuth:      true,
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.ShareTextHandler),
		},
	}
}

func (r *Router) storeName(isPermanent bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		storage := mux.Vars(request)["storage"]
		if storage == "" {
			log.Printf("Invalid storage request, url = %s", request.URL)
			next.ServeHTTP(w, request)
			return
		}

		context.Set(request, ContextStorageName, storage)
		if isPermanent {
			context.Set(request, ContextStorageIsPermanent, true)
		}

		next.ServeHTTP(w, request)
	})
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

		if route.StorePath || route.PermanentPath {
			handler = r.storeName(route.PermanentPath, handler)
		}

		handler = r.h.RecoverHandler(handler)

		muxRoute.Handler(handler)
	}

	return router
}
