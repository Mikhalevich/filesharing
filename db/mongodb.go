package db

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DatabaseName           = "users"
	DatabaseHost           = "localhost"
	CollectionUsers        = "users"
	CollectionLoginRequest = "request"
)

var (
	sessionPool *mgo.Session
)

func init() {
	var err error
	if sessionPool, err = mgo.Dial(DatabaseHost); err != nil {
		panic(err)
	}

	storage := NewStorage()
	if err = storage.createIndexes(); err != nil {
		panic(err)
	}

	if err = storage.clearTemporaryData(); err != nil {
		panic(err)
	}
}

type Storage struct {
	session *mgo.Session
}

func NewStorage() *Storage {
	storage := &Storage{
		session: sessionPool.Copy(),
	}

	return storage
}

func (self *Storage) cUsers() *mgo.Collection {
	return self.session.DB(DatabaseName).C(CollectionUsers)
}

func (self *Storage) cLoginRequest() *mgo.Collection {
	return self.session.DB(DatabaseName).C(CollectionLoginRequest)
}

func (self *Storage) Close() {
	self.session.Close()
}

func (self *Storage) createIndexes() error {
	userIndex := mgo.Index{
		Key:      []string{"name"},
		Unique:   true,
		DropDups: true,
	}
	if err := self.cUsers().EnsureIndex(userIndex); err != nil {
		return err
	}

	loginRequestIndex := mgo.Index{
		Key:      []string{"name", "remote_addr"},
		Unique:   true,
		DropDups: true,
	}
	if err := self.cLoginRequest().EnsureIndex(loginRequestIndex); err != nil {
		return err
	}

	return nil
}

func (self *Storage) clearTemporaryData() error {
	return self.ClearRequests()
}

func (self *Storage) GetRequest(name, remoteAddr string) (LoginRequest, error) {
	request := LoginRequest{}
	if err := self.cLoginRequest().Find(bson.M{"name": name, "remote_addr": remoteAddr}).One(&request); err != nil {
		return LoginRequest{}, err
	}

	return request, nil
}

func (self *Storage) AddRequest(name, remoteAddr string) error {
	// try to find login request first
	request := LoginRequest{}
	if err := self.cLoginRequest().Find(bson.M{"name": name, "remote_addr": remoteAddr}).One(&request); err == nil {
		// request exists
		request.Id = ""
		request.LastRequest = time.Now().Unix()
		request.Count = request.Count + 1

		if err := self.cLoginRequest().Update(bson.M{"name": name, "remote_addr": remoteAddr}, request); err != nil {
			return err
		}
	} else {
		// new reqeust
		request.UserName = name
		request.RemoteAddr = remoteAddr
		request.LastRequest = time.Now().Unix()
		request.Count = 1
		if err := self.cLoginRequest().Insert(request); err != nil {
			return err
		}
	}

	return nil
}

func (self *Storage) RemoveRequest(name, remoteAddr string) error {
	return self.cLoginRequest().Remove(bson.M{"name": name, "remote_addr": remoteAddr})
}

func (self *Storage) ResetRequestCounter(request LoginRequest) error {
	request.Id = ""
	request.Count = 1
	return self.cLoginRequest().Update(bson.M{"name": request.UserName, "remote_addr": request.RemoteAddr}, request)
}

func (self *Storage) ClearRequests() error {
	_, err := self.cLoginRequest().RemoveAll(bson.M{})
	return err
}

func (self *Storage) UserByName(name string) (User, error) {
	user := User{}
	err := self.cUsers().Find(bson.M{"name": name}).One(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (self *Storage) UserByNameAndPassword(name string, password TypePassword) (User, error) {
	user := User{}
	err := self.cUsers().Find(bson.M{"name": name, "password": password}).One(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (self *Storage) UserBySessionId(sessionId string) (User, error) {
	user := User{}
	if err := self.cUsers().Find(bson.M{"session.id": sessionId}).One(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (self *Storage) AddUser(user *User) error {
	if err := self.cUsers().Insert(user); err != nil {
		return err
	}

	return nil
}

func (self *Storage) AddSession(id bson.ObjectId, sessionId string, expires int64) error {
	return self.cUsers().UpdateId(id, bson.M{"$push": bson.M{"sessions": bson.M{"id": sessionId, "expires": expires}}})
}

func (self *Storage) RemoveExpiredSessions(id bson.ObjectId, checkTime int64) error {
	return self.cUsers().UpdateId(id, bson.M{"$pull": bson.M{"sessions": bson.M{"expires": bson.M{"$lt": checkTime}}}})
}
