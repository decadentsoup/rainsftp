package main

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
)

type ldapAuthBackend struct {
	log        *logrus.Logger
	endpoint   string
	username   string
	password   string
	baseDN     string
	readGroup  string
	writeGroup string
}

func newLDAPAuthBackend(log *logrus.Logger) *ldapAuthBackend {
	backend := &ldapAuthBackend{
		log:        log,
		endpoint:   os.Getenv("LDAP_ENDPOINT"),
		username:   os.Getenv("LDAP_USERNAME"),
		password:   os.Getenv("LDAP_PASSWORD"),
		baseDN:     os.Getenv("LDAP_BASE_DN"),
		readGroup:  os.Getenv("LDAP_READ_GROUP"),
		writeGroup: os.Getenv("LDAP_WRITE_GROUP"),
	}

	log.WithFields(logrus.Fields{
		"endpoint":   backend.endpoint,
		"username":   backend.username,
		"password":   "[REDACTED]",
		"baseDN":     backend.baseDN,
		"readGroup":  backend.readGroup,
		"writeGroup": backend.writeGroup,
	}).Info("testing ldap credentials...")

	if ldapClient := backend.dial(); ldapClient == nil {
		os.Exit(1)
	} else {
		ldapClient.Close()
	}

	log.Info("tested ldap credentials")

	return backend
}

func (backend *ldapAuthBackend) auth(username string, password string) *permissions {
	ldapClient := backend.dial()

	if ldapClient == nil {
		backend.log.Error("cannot authenticate")
		return nil
	}

	defer ldapClient.Close()

	result, err := ldapClient.Search(
		ldap.NewSearchRequest(
			backend.baseDN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0,
			0,
			false,
			fmt.Sprintf(
				"(&(objectClass=organizationalPerson)(uid=%s))",
				ldap.EscapeFilter(username),
			),
			[]string{"memberOf"},
			nil,
		),
	)

	if err != nil {
		backend.log.WithError(err).Error("failed to search ldap")
		return nil
	}

	entryCount := len(result.Entries)

	if entryCount == 0 {
		backend.log.Info("username not found")
		return nil
	}

	if entryCount != 1 {
		backend.log.WithField("entryCount", entryCount).Error("ldap returned suspicious number of entries")
		return nil
	}

	if err := ldapClient.Bind(result.Entries[0].DN, password); err != nil {
		backend.log.WithError(err).Info("authentication failed")
		return nil
	}

	permissions := permissions{canRead: false, canWrite: false}

	for _, group := range result.Entries[0].GetAttributeValues("memberOf") {
		switch group {
		case backend.readGroup:
			permissions.canRead = true
		case backend.writeGroup:
			permissions.canWrite = true
		}
	}

	return &permissions
}

func (backend *ldapAuthBackend) dial() *ldap.Conn {
	ldapClient, err := ldap.DialURL(backend.endpoint)

	if err != nil {
		backend.log.WithError(err).Error("failed to connect to LDAP")
		return nil
	}

	if url, err := url.Parse(backend.endpoint); err != nil {
		backend.log.WithError(err).Error("failed to parse LDAP_ENDPOINT")
		ldapClient.Close()
		return nil
	} else if url.Scheme != "ldaps" {
		backend.log.Info("securing ldap connection...")

		if err := ldapClient.StartTLS(&tls.Config{ServerName: url.Hostname()}); err != nil {
			backend.log.WithError(err).Error("failed to secure connection to LDAP")
			ldapClient.Close()
			return nil
		}
	}

	backend.log.Info("ldap connection secured")

	backend.log.Infof("what %v %v", backend.username, backend.password)
	if err := ldapClient.Bind(backend.username, backend.password); err != nil {
		backend.log.WithError(err).Error("failed to bind to LDAP")
		ldapClient.Close()
		return nil
	}

	return ldapClient
}
