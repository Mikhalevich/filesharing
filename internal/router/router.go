package router

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Mikhalevich/filesharing/pkg/ctxinfo"
)

type route struct {
	Pattern string
	Methods string
	Public  bool
	Handler http.Handler
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

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

func configure(h handler) []route {
	return []route{
		{
			Pattern: "/register/",
			Methods: "POST",
			Public:  true,
			Handler: http.HandlerFunc(h.RegisterHandler),
		},
		{
			Pattern: "/login/",
			Methods: "POST",
			Public:  true,
			Handler: http.HandlerFunc(h.LoginHandler),
		},
		{
			Pattern: "/index.html",
			Methods: "GET",
			Handler: http.HandlerFunc(h.IndexHTMLHandler),
		},
		{
			Pattern: "/file/",
			Methods: "GET",
			Handler: http.HandlerFunc(h.GetFileHandler),
		},
		{
			Pattern: "/list/",
			Methods: "GET",
			Handler: http.HandlerFunc(h.GetFileList),
		},
		{
			Pattern: "/upload/",
			Methods: "POST",
			Handler: http.HandlerFunc(h.UploadHandler),
		},
		{
			Pattern: "/remove/",
			Methods: "POST",
			Handler: http.HandlerFunc(h.RemoveHandler),
		},
		{
			Pattern: "/shareText/",
			Methods: "POST",
			Handler: http.HandlerFunc(h.ShareTextHandler),
		},
	}
}

func storeParametes(isPublic bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		storage := r.FormValue("storage")
		if storage != "" {
			ctx = ctxinfo.WithUserName(ctx, storage)

			permanent := r.FormValue("permanent")
			if permanent != "" {
				ctx = ctxinfo.WithPermanentStorage(ctx, true)
			}
		}

		ctx = ctxinfo.WithPublicStorage(ctx, isPublic)

		fileName := r.FormValue("file")
		if fileName != "" {
			ctx = ctxinfo.WithFileName(ctx, fileName)
		}
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func MakeRoutes(router *mux.Router, authEnabled bool, h handler, l Logger) {
	for _, route := range configure(h) {
		muxRoute := router.NewRoute()
		muxRoute.Path(route.Pattern)

		muxRoute.Methods(strings.Split(route.Methods, ",")...)

		handler := route.Handler
		if authEnabled && !route.Public {
			handler = h.CheckAuthMiddleware(handler)
		}

		handler = h.CreateStorageMiddleware(handler)

		handler = storeParametes(route.Public, handler)

		handler = h.RecoverMiddleware(handler)

		muxRoute.Handler(handler)
	}
}
