package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
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

func Run(name string, cfg Configer, setup func(srv micro.Service, s Servicer) error) {
	l := newLoggerWrapper(name)

	if name == "" {
		l.Error("service name is empty")
		return
	}

	if err := loadConfig(cfg, os.Getenv("FS_CONFIG_FILE")); err != nil {
		l.WithError(err).Error("load config error")
		return
	}

	serviceCfg := cfg.Service()

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

	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%d", serviceCfg.Port),
		Handler: srvOptions.router,
	}

	go func() {
		l.Infof("http server started at %d", serviceCfg.Port)
		defer l.Info("http server stopped")
		if err := httpServer.ListenAndServe(); err != nil {
			l.WithError(err).Error("failed to run http server")
			return
		}
	}()

	l.Info("rpc server started")
	if err := srv.Run(); err != nil {
		l.WithError(err).Error("failed to run service")
		return
	}
	l.Info("rpc server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		l.WithError(err).Error("failed to shutdown http server")
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
