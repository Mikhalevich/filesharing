package handlers

import (
	"crypto/sha1"
	"time"

	"github.com/Mikhalevich/filesharing/db"
)

func crypt(password string) [sha1.Size]byte {
	if password != "" {
		return sha1.Sum([]byte(password))
	}
	return [sha1.Size]byte{}
}

func addUser(name string, password string, s *db.Session) error {
	storage := db.NewStorage()
	defer storage.Close()

	user := &db.User{
		Name:     name,
		Password: crypt(password),
		Sessions: []db.Session{*s},
	}

	return storage.AddUser(user)
}

func sessionByUserName(name string, id string) (*db.Session, error) {
	storage := db.NewStorage()
	defer storage.Close()

	user, err := storage.UserByName(name)
	if err != nil {
		return nil, err
	}

	session, err := user.SessionById(id)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func requestWaitPeriod(name string, host string, maxCount int, waitPeriod int64) (int64, error) {
	storage := db.NewStorage()
	defer storage.Close()

	loginRequest, err := storage.GetRequest(name, host)
	if err != nil {
		return 0, err
	}

	if loginRequest.Count >= maxCount {
		timeDelta := time.Now().Unix() - loginRequest.LastRequest
		allowed := timeDelta >= waitPeriod

		if !allowed {
			return waitPeriod - timeDelta, nil
		}

		err = storage.ResetRequestCounter(loginRequest)
		if err != nil {
			return 0, err
		}
	}

	return 0, nil
}
