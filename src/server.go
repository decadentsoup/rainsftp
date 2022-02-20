package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/go-ldap/ldap/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	gossh "golang.org/x/crypto/ssh"
)

func main() {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if debug, err := strconv.ParseBool(getEnvWithDefault("DEBUG", "false")); err != nil {
		log.Panic("environment variable DEBUG has invalid value (must be \"true\" or \"false\")")
	} else if debug {
		log.SetLevel(logrus.TraceLevel)
	} else {
		log.SetFormatter(&ecslogrus.Formatter{})
	}

	ldapEndpoint := os.Getenv("LDAP_ENDPOINT")
	ldapBaseDN := os.Getenv("LDAP_BASE_DN")
	ldapReadGroup := os.Getenv("LDAP_READ_GROUP")
	ldapWriteGroup := os.Getenv("LDAP_WRITE_GROUP")

	log.WithFields(logrus.Fields{
		"ldapEndpoint":   ldapEndpoint,
		"ldapBaseDN":     ldapBaseDN,
		"ldapReadGroup":  ldapReadGroup,
		"ldapWriteGroup": ldapWriteGroup,
	}).Info("testing ldap credentials...")

	if ldapClient := dialLDAP(log.WithFields(logrus.Fields{})); ldapClient == nil {
		os.Exit(1)
	} else {
		ldapClient.Close()
	}

	log.Info("tested ldap credentials")

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Secure, err := strconv.ParseBool(getEnvWithDefault("S3_SECURE", "true"))
	if err != nil {
		log.Panic("environment variable S3_SECURE has invalid value (must be \"true\" or \"false\")")
	}
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3Secret := os.Getenv("S3_SECRET")
	s3Bucket := os.Getenv("S3_BUCKET")

	log.WithFields(logrus.Fields{
		"endpoint": s3Endpoint,
		"secure":   s3Secure,
		"bucket":   s3Bucket,
	}).Info("connecting to storage service...")

	minioClient, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3AccessKey, s3Secret, ""),
		Secure: s3Secure,
	})

	if err != nil {
		log.WithError(err).Panic("failed to connect to storage bucket")
	}

	log.Info("connected to storage service")

	log.Info("initializing ssh server...")

	sshServer := &ssh.Server{
		Addr: ":8022",
		Handler: func(session ssh.Session) {
			log.WithField("address", session.RemoteAddr().String()).Info("client attempted connection without sftp")
			io.WriteString(session, "This server only supports SFTP.\n")
		},
		PasswordHandler: func(context ssh.Context, password string) bool {
			username := context.User()
			log := log.WithField("address", context.RemoteAddr().String()).WithField("username", username)
			log.Info("authenticating...")

			if ldapClient := dialLDAP(log); ldapClient == nil {
				log.Error("cannot authenticate")
				return false
			} else {
				defer ldapClient.Close()

				if result, err := ldapClient.Search(ldap.NewSearchRequest(ldapBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, fmt.Sprintf("(&(objectClass=organizationalPerson)(uid=%s))", ldap.EscapeFilter(username)), []string{"memberOf"}, nil)); err != nil {
					log.WithError(err).Error("failed to search ldap")
					return false
				} else if entryCount := len(result.Entries); entryCount == 0 {
					log.Info("username not found")
					return false
				} else if entryCount != 1 {
					log.WithField("entryCount", entryCount).Error("ldap returned suspicious number of entries")
				} else if err := ldapClient.Bind(result.Entries[0].DN, password); err != nil {
					log.WithError(err).Info("authentication failed")
					return false
				} else {
					context.SetValue("AllowRead", false)
					context.SetValue("AllowWrite", false)

					for _, group := range result.Entries[0].GetAttributeValues("memberOf") {
						switch group {
						case ldapReadGroup:
							context.SetValue("AllowRead", true)
						case ldapWriteGroup:
							context.SetValue("AllowWrite", true)
						}
					}

					return true
				}
			}

			return false
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": func(session ssh.Session) {
				log := log.WithField("address", session.RemoteAddr().String())
				log.Info("client connected")
				defer log.Info("client disconnected")

				context := session.Context()
				allowRead := context.Value("AllowRead").(bool)
				allowWrite := context.Value("AllowWrite").(bool)

				log.WithFields(logrus.Fields{
					"allowRead":  allowRead,
					"allowWrite": allowWrite,
				}).Info("permissions")

				handler := &handler{log, minioClient, s3Bucket, allowRead, allowWrite}

				sftpServer := sftp.NewRequestServer(session, sftp.Handlers{
					FileGet:  handler,
					FilePut:  handler,
					FileCmd:  handler,
					FileList: handler,
				})

				if err := sftpServer.Serve(); err != nil && err != io.EOF {
					log.WithError(err).Error("failed to serve sftp")
				}
			},
		},
	}

	log.Info("initialized ssh server")

	log.Info("registering ssh host keys...")

	for _, key := range strings.Split(os.Getenv("HOST_KEYS"), ",") {
		if signer, err := gossh.ParsePrivateKey([]byte(key)); err != nil {
			log.WithError(err).Panic("failed to parse host key")
		} else {
			sshServer.AddHostKey(signer)
		}
	}

	log.Info("registered ssh host keys")

	log.WithField("address", sshServer.Addr).Info("ready")
	log.Fatal(sshServer.ListenAndServe())
}

func getEnvWithDefault(key string, fallback string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		return fallback
	}

	return value
}
