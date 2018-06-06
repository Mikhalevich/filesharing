package db

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

type Null struct {
	// pass
}

func NewNullStorage() *Null {
	return &Null{}
}

func (n *Null) Close() {
	// pass
}

func (n *Null) GetRequest(name, remoteAddr string) (LoginRequest, error) {
	return LoginRequest{}, errors.New("Null impl")
}

func (n *Null) AddRequest(name, remoteAddr string) error {
	return errors.New("Null impl")
}

func (n *Null) RemoveRequest(name, remoteAddr string) error {
	return errors.New("Null impl")
}

func (n *Null) ResetRequestCounter(request LoginRequest) error {
	return errors.New("Null impl")
}

func (n *Null) ClearRequests() error {
	return errors.New("Null impl")
}

func (n *Null) UserByName(name string) (User, error) {
	return User{}, errors.New("Null impl")
}

func (n *Null) UserByNameAndPassword(name string, password TypePassword) (User, error) {
	return User{}, errors.New("Null impl")
}

func (n *Null) UserBySessionId(sessionId string) (User, error) {
	return User{}, errors.New("Null impl")
}

func (n *Null) AddUser(user *User) error {
	return errors.New("Null impl")
}

func (n *Null) AddSession(id bson.ObjectId, sessionId string, expires int64) error {
	return errors.New("Null impl")
}

func (n *Null) RemoveExpiredSessions(id bson.ObjectId, checkTime int64) error {
	return errors.New("Null impl")
}
