package router

import (
	"net/http"
	"strings"

	"github.com/Mikhalevich/filesharing/ctxinfo"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Route struct {
	Pattern       string
	IsPrefix      bool
	Methods       string
	Public        bool
	PermanentPath bool
	Handler       http.Handler
}

type handler interface {
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	GetFileList(w http.ResponseWriter, r *http.Request)
	IndexHTMLHandler(w http.ResponseWriter, r *http.Request)
	UploadHandler(w http.ResponseWriter, r *http.Request)
	RemoveHandler(w http.ResponseWriter, r *http.Request)
	GetFileHandler(w http.ResponseWriter, r *http.Request)
	ShareTextHandler(w http.ResponseWriter, r *http.Request)
	CheckAuthMiddleware(next http.Handler) http.Handler
	CreateStorageMiddleware(next http.Handler) http.Handler
	RecoverMiddleware(next http.Handler) http.Handler
}

type Router struct {
	enableAuth bool
	routes     []Route
	h          handler
	logger     *logrus.Logger
}

func NewRouter(ea bool, handl handler, l *logrus.Logger) *Router {
	return &Router{
		enableAuth: ea,
		h:          handl,
		logger:     l,
	}
}

func (r *Router) makeRoutes() {
	r.routes = []Route{
		{
			Pattern: "/common/",
			Methods: "GET",
			Public:  true,
			Handler: http.HandlerFunc(r.h.GetFileList),
		},
		{
			Pattern: "/register/",
			Methods: "GET,POST",
			Public:  true,
			Handler: http.HandlerFunc(r.h.RegisterHandler),
		},
		{
			Pattern: "/login/{storage}/",
			Methods: "GET,POST",
			Public:  true,
			Handler: http.HandlerFunc(r.h.LoginHandler),
		},
		{
			Pattern: "/{storage}/index.html",
			Methods: "GET",
			Handler: http.HandlerFunc(r.h.IndexHTMLHandler),
		},
		{
			Pattern:       "/{storage}/permanent/index.html",
			Methods:       "GET",
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.IndexHTMLHandler),
		},
		{
			Pattern:       "/{storage}/permanent/{file}/",
			Methods:       "GET",
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.GetFileHandler),
		},
		{
			Pattern:       "/{storage}/permanent/",
			Methods:       "GET",
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.GetFileList),
		},
		{
			Pattern: "/{storage}/{file}/",
			Methods: "GET",
			Handler: http.HandlerFunc(r.h.GetFileHandler),
		},
		{
			Pattern: "/{storage}/",
			Methods: "GET",
			Handler: http.HandlerFunc(r.h.GetFileList),
		},
		{
			Pattern: "/{storage}/upload/",
			Methods: "POST",
			Handler: http.HandlerFunc(r.h.UploadHandler),
		},
		{
			Pattern:       "/{storage}/permanent/upload/",
			Methods:       "POST",
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.UploadHandler),
		},
		{
			Pattern: "/{storage}/remove/",
			Methods: "POST",
			Handler: http.HandlerFunc(r.h.RemoveHandler),
		},
		{
			Pattern:       "/{storage}/permanent/remove/",
			Methods:       "POST",
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.RemoveHandler),
		},
		{
			Pattern: "/{storage}/shareText/",
			Methods: "POST",
			Handler: http.HandlerFunc(r.h.ShareTextHandler),
		},
		{
			Pattern:       "/{storage}/permanent/shareText/",
			Methods:       "POST",
			PermanentPath: true,
			Handler:       http.HandlerFunc(r.h.ShareTextHandler),
		},
	}
}

func (r *Router) storeRouterParametes(isPublic bool, isPermanent bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		ctx := ctxinfo.WithPublicStorage(request.Context(), isPublic)

		storage := mux.Vars(request)["storage"]
		if storage != "" {
			ctx = ctxinfo.WithUserName(ctx, storage)
			ctx = ctxinfo.WithPermanentStorage(ctx, isPermanent)
		}

		fileName := mux.Vars(request)["file"]
		if fileName != "" {
			ctx = ctxinfo.WithFileName(ctx, fileName)
		}
		request = request.WithContext(ctx)

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
		if r.enableAuth && !route.Public {
			handler = r.h.CheckAuthMiddleware(handler)
		}

		handler = r.h.CreateStorageMiddleware(handler)

		handler = r.storeRouterParametes(route.Public, route.PermanentPath, handler)

		handler = r.h.RecoverMiddleware(handler)

		muxRoute.Handler(handler)
	}

	return router
}
