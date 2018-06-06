package db

import (
	"crypto/sha1"
	"errors"

	"gopkg.in/mgo.v2/bson"
)

var (
	UseDB bool = true
)

type TypePassword [sha1.Size]byte

func (self TypePassword) IsEmpty() bool {
	for _, value := range self {
		if value != 0 {
			return false
		}
	}
	return true
}

type LoginRequest struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	UserName    string        `bson:"name"`
	RemoteAddr  string        `bson:"remote_addr"`
	LastRequest int64         `bson:"last_request"`
	Count       int           `bson:"count"`
}

type Session struct {
	Id      string `bson:"id"`
	Expires int64  `bson:"expires"`
}

type User struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Name     string        `bson:"name"`
	Password TypePassword  `bson:"password"`
	Sessions []Session     `bson:"sessions"`
}

func (self User) SessionById(id string) (Session, error) {
	for _, session := range self.Sessions {
		if session.Id == id {
			return session, nil
		}
	}

	return Session{}, errors.New("Not found")
}

type Storager interface {
	Close()
	GetRequest(name, remoteAddr string) (LoginRequest, error)
	AddRequest(name, remoteAddr string) error
	RemoveRequest(name, remoteAddr string) error
	ResetRequestCounter(request LoginRequest) error
	ClearRequests() error
	UserByName(name string) (User, error)
	UserByNameAndPassword(name string, password TypePassword) (User, error)
	UserBySessionId(sessionId string) (User, error)
	AddUser(user *User) error
	AddSession(id bson.ObjectId, sessionId string, expires int64) error
	RemoveExpiredSessions(id bson.ObjectId, checkTime int64) error
}

func NewStorage() Storager {
	if UseDB {
		return NewMgoStorage()
	}

	return NewNullStorage()
}
