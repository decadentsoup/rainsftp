package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

type authBackend interface {
	auth(username string, password string) *permissions
}

type permissions struct {
	canRead  bool
	canWrite bool
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
