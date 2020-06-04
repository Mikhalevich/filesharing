package main

import (
	"net/http"
	"time"

	"github.com/Mikhalevich/filesharing/handlers"
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

func (cs *CookieSession) GetToken(r *http.Request) (*handlers.Token, error) {
	name := cs.namer.Name(r)
	for _, cook := range r.Cookies() {
		if cook.Name != name {
			continue
		}

		if cook.Expires.Unix() < time.Now().Unix() {
			return nil, handlers.ErrExpired
		}

		return &handlers.Token{
			Value: cook.Value,
		}, nil
	}

	return nil, handlers.ErrNotExist
}

func (cs *CookieSession) SetToken(w http.ResponseWriter, token *handlers.Token, name string) {
	cookie := http.Cookie{Name: name, Value: token.Value, Path: "/", Expires: time.Now().Add(time.Duration(cs.expirePeriod) * time.Second), HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func (cs *CookieSession) Remove(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
}

// func (cs *CookieSession) Create() goauth.Session {
// 	bytes := make([]byte, 32)
// 	rand.Read(bytes)
// 	return *goauth.NewSession("session", base64.URLEncoding.EncodeToString(bytes), cs.expirePeriod)
// }
