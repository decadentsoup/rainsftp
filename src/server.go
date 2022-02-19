package main

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
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
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": func(session ssh.Session) {
				log := log.WithField("address", session.RemoteAddr().String())
				log.Info("client connected")
				defer log.Info("client disconnected")

				handler := handler{log, minioClient, s3Bucket}

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
