package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
)

type publicKeyAuthBackend struct {
	log            *logrus.Logger
	users          map[string]publicKeyUser
	authorizedKeys []byte
}

type publicKeyUser struct {
	PublicKey string
	CanRead   bool
	CanWrite  bool
}

func newPublicKeyBackend(log *logrus.Logger) *publicKeyAuthBackend {
	backend := &publicKeyAuthBackend{
		log:   log,
		users: make(map[string]publicKeyUser),
	}

	if err := json.Unmarshal([]byte(os.Getenv("PUBLIC_KEY_USERS")), &backend.users); err != nil {
		log.WithError(err).Error("failed to parse PUBLIC_KEY_USERS")
	}

	authorizedKeyStrings := []string{}
	for _, user := range backend.users {
		authorizedKeyStrings = append(authorizedKeyStrings, user.PublicKey)
	}

	backend.authorizedKeys = []byte(strings.Join(authorizedKeyStrings, "\n"))

	return backend
}

func (backend *publicKeyAuthBackend) auth(context ssh.Context, publicKey ssh.PublicKey) *permissions {
	data := backend.authorizedKeys
	allowed, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		log.Println("Failed to Parse Key")
		log.Println(err.Error())
	}

	if user, ok := backend.users[context.User()]; ok {
		if !ssh.KeysEqual(publicKey, allowed) {
			return nil
		}
		return &permissions{
			canRead:  user.CanRead,
			canWrite: user.CanWrite,
		}
	}

	return nil

}
