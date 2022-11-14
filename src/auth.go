package main

import (
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
)

type authBackend interface {
	auth(username string, password string) *permissions
}
type authKeyBackend interface {
	auth(context ssh.Context, publicKey ssh.PublicKey) *permissions
}

type permissions struct {
	canRead  bool
	canWrite bool
}

func newAuthKeyBackend(log *logrus.Logger) authKeyBackend {
	if os.Getenv("PUBLIC_KEY_USERS") != "" {
		log.Info("using public key users found, adding public key backend")
		return newPublicKeyBackend(log)
	}
	log.Println("no public key users found, not using public key auth")
	return nil
}

func newAuthBackend(log *logrus.Logger) authBackend {
	if os.Getenv("JSON_USERS") != "" {
		log.Info("using json authentication backend")
		return newJSONAuthBackend(log)
	}

	if os.Getenv("LDAP_ENDPOINT") != "" {
		log.Info("using ldap authentication backend")
		return newLDAPAuthBackend(log)
	}

	log.Error("no authentication mechanism provided")
	os.Exit(1)
	return nil
}
