package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/server"
	"github.com/sirupsen/logrus"
)

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

type Servicer interface {
	Logger() Logger
}

type microService struct {
	srv micro.Service
	l   *logrus.Logger
}

func New(name string) (*microService, error) {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.JSONFormatter{})

	f, err := os.OpenFile(fmt.Sprintf("/log/%s.log", name), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	l.SetOutput(io.MultiWriter(os.Stdout, f))

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
			defer l.Infof("end processing %s, time = %v", req.Method(), time.Now().Sub(start))
			err := fn(ctx, req, rsp)
			if err != nil {
				l.Errorf("processing %s error: %v", req.Method(), err)
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
		ms.l.Errorf("register handler error: %v\n", err)
		return err
	}
	return nil
}

func (ms *microService) Run() error {
	if err := ms.srv.Run(); err != nil {
		ms.l.Errorf("run micro service error: %v\n", err)
		return err
	}
	return nil
}
