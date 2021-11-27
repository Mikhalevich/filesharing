package service

import (
	"context"
	"time"

	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/server"
)

type Servicer interface {
	Logger() Logger
}

type microService struct {
	srv micro.Service
	l   Logger
}

func New(name string) (*microService, error) {
	l := newLoggerWrapper(name)

	srv := micro.NewService(
		micro.Name(name),
		micro.WrapHandler(makeLoggerWrapper(l)),
	)

	srv.Init()

	return &microService{
		srv: srv,
		l:   l,
	}, nil
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

func (ms *microService) Logger() Logger {
	return ms.l
}

func (ms *microService) RegisterHandler(fn func(srv micro.Service, s Servicer) error) error {
	if err := fn(ms.srv, ms); err != nil {
		ms.l.WithError(err).Error("failed to register handler")
		return err
	}
	return nil
}

func (ms *microService) Run() error {
	if err := ms.srv.Run(); err != nil {
		ms.l.WithError(err).Error("failed to run service")
		return err
	}
	return nil
}
