package main

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type jsonAuthBackend struct {
	log   *logrus.Logger
	users map[string]jsonUser
}

type jsonUser struct {
	Password string
	CanRead  bool
	CanWrite bool
}

func newJSONAuthBackend(log *logrus.Logger) *jsonAuthBackend {
	backend := &jsonAuthBackend{
		log:   log,
		users: make(map[string]jsonUser),
	}

	if err := json.Unmarshal([]byte(os.Getenv("JSON_USERS")), &backend.users); err != nil {
		log.WithError(err).Error("failed to parse JSON_USERS")
		os.Exit(1)
	}

	return backend
}

func (backend *jsonAuthBackend) auth(username string, password string) *permissions {
	if user, ok := backend.users[username]; ok {
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			return nil
		}

		return &permissions{
			canRead:  user.CanRead,
			canWrite: user.CanWrite,
		}
	}

	return nil
}
