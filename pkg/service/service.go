package service

import (
	"github.com/gorilla/mux"
)

type Option func(o *service)

func WithPostAction(fn func()) Option {
	return func(o *service) {
		o.postActions = append(o.postActions, fn)
	}
}

type service struct {
	l           Logger
	router      *mux.Router
	postActions []func()
}

func (s *service) Logger() Logger {
	return s.l
}

func (s *service) Router() *mux.Router {
	return s.router
}

func (s *service) AddOption(opt Option) {
	opt(s)
}

func (s *service) runPostActions() {
	for _, action := range s.postActions {
		action()
	}
}
