package main

import (
	"crypto/tls"
	"net/url"
	"os"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
)

func dialLDAP(log *logrus.Entry) *ldap.Conn {
	ldapEndpoint := os.Getenv("LDAP_ENDPOINT")
	ldapClient, err := ldap.DialURL(ldapEndpoint)

	if err != nil {
		log.WithError(err).Error("failed to connect to LDAP")
		return nil
	}

	if url, err := url.Parse(ldapEndpoint); err != nil {
		log.WithError(err).Error("failed to parse LDAP_ENDPOINT")
		ldapClient.Close()
		return nil
	} else if url.Scheme != "ldaps" {
		log.Info("securing ldap connection...")

		if err := ldapClient.StartTLS(&tls.Config{ServerName: url.Hostname()}); err != nil {
			log.WithError(err).Error("failed to secure connection to LDAP")
			ldapClient.Close()
			return nil
		}
	}

	log.Info("ldap connection secured")

	if err := ldapClient.Bind(os.Getenv("LDAP_USERNAME"), os.Getenv("LDAP_PASSWORD")); err != nil {
		log.WithError(err).Error("failed to bind to LDAP")
		ldapClient.Close()
		return nil
	}

	return ldapClient
}
