package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/Mikhalevich/goauth"
)

type Namer interface {
	Name(r *http.Request) string
}

type CookieSession struct {
	namer        Namer
	expirePeriod int64
}

func NewCookieSession(n Namer, period int64) *CookieSession {
	return &CookieSession{
		namer:        n,
		expirePeriod: period,
	}
}

func (cs *CookieSession) Create() goauth.Session {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return *goauth.NewSession("session", base64.URLEncoding.EncodeToString(bytes), cs.expirePeriod)
}

func (cs *CookieSession) Find(r *http.Request) (goauth.Session, error) {
	name := cs.namer.Name(r)
	for _, cook := range r.Cookies() {
		if cook.Name != name {
			continue
		}

		return goauth.Session{
			Name:    name,
			Value:   cook.Value,
			Expires: cook.Expires.Unix(),
		}, nil
	}

	return goauth.Session{}, goauth.ErrNotExists
}
