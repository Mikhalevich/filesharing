package service

import (
	"context"
	"net/http"
	"time"

	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/server"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Servicer interface {
	Logger() Logger
	Router() *mux.Router
	AddOption(opt Option)
}

func Run(name string, host string, setup func(srv micro.Service, s Servicer) error) {
	l := newLoggerWrapper(name)

	srv := micro.NewService(
		micro.Name(name),
		micro.WrapHandler(makeLoggerWrapper(l)),
	)

	srv.Init()

	srvOptions := service{
		l:      l,
		router: mux.NewRouter().StrictSlash(true),
	}

	srvOptions.router.Path("/metrics/").Handler(promhttp.Handler())

	if err := setup(srv, &srvOptions); err != nil {
		l.WithError(err).Error("failed to setup service")
		return
	}

	defer srvOptions.runPostActions()

	if host != "" {
		go func() {
			l.Info("http server started")
			defer l.Info("http server stopped")
			if err := http.ListenAndServe(host, srvOptions.router); err != nil {
				l.WithError(err).Error("failed to run http server")
			}
		}()
	}

	l.Info("server started")
	if err := srv.Run(); err != nil {
		l.WithError(err).Error("failed to run service")
		return
	}
}

func makeLoggerWrapper(l Logger) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			l.Infof("processing %s", req.Method())
			start := time.Now()
			defer l.Infof("end processing %s, time = %v", req.Method(), time.Since(start))
			err := fn(ctx, req, rsp)
			if err != nil {
				l.WithError(err).WithFields(map[string]interface{}{
					"method": req.Method(),
				}).Error("failed to execute handler")
			}
			return err
		}
	}
}
