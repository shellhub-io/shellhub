package connector

import (
	"archive/tar"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"

	cerrdefs "github.com/containerd/errdefs"
	dockerclient "github.com/docker/docker/client"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/agent/server/modes"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

var _ modes.Authenticator = (*Authenticator)(nil)

// Authenticator implements the Authenticator interface when the server is running in connector mode.
type Authenticator struct {
	api       client.Client
	authData  *models.DeviceAuthResponse
	container *string
	docker    dockerclient.APIClient
}

// NewAuthenticator creates a new instance of Authenticator for the connector mode.
func NewAuthenticator(api client.Client, docker dockerclient.APIClient, authData *models.DeviceAuthResponse, container *string) *Authenticator {
	return &Authenticator{
		api:       api,
		authData:  authData,
		container: container,
		docker:    docker,
	}
}

type containerFile struct {
	io.Reader
	io.Closer
}

func readContainerFile(ctx context.Context, cli dockerclient.APIClient, container, path string) (io.ReadCloser, error) {
	archive, _, err := cli.CopyFromContainer(ctx, container, path)
	if err != nil {
		return nil, err
	}

	file := tar.NewReader(archive)
	if _, err := file.Next(); err != nil {
		archive.Close() //nolint:errcheck // the read already failed, and that error is the one returned

		return nil, err
	}

	return containerFile{Reader: file, Closer: archive}, nil
}

func getPasswd(ctx context.Context, cli dockerclient.APIClient, container string) (io.ReadCloser, error) {
	return readContainerFile(ctx, cli, container, "/etc/passwd")
}

func getShadow(ctx context.Context, cli dockerclient.APIClient, container string) (io.ReadCloser, error) {
	return readContainerFile(ctx, cli, container, "/etc/shadow")
}

func accountExpiredInContainer(ctx context.Context, cli dockerclient.APIClient, container, username string) bool {
	shadow, err := getShadow(ctx, cli, container)
	if cerrdefs.IsNotFound(err) {
		if _, inspectErr := cli.ContainerInspect(ctx, container); inspectErr != nil {
			log.WithFields(
				log.Fields{
					"container": container,
					"username":  username,
				},
			).WithError(inspectErr).Error("failed to inspect the container missing a shadow file")

			return true
		}

		return false
	}

	if err != nil {
		log.WithFields(
			log.Fields{
				"container": container,
				"username":  username,
			},
		).WithError(err).Error("failed to get the shadow file from container")

		return true
	}
	defer shadow.Close() //nolint:errcheck // response body of a finished read, nothing to flush

	return osauth.AccountExpiredFromShadow(username, shadow)
}

// Password handles the server's SSH password authentication when server is running in connector mode.
func (a *Authenticator) Password(ctx gliderssh.Context, username string, password string) bool {
	passwd, err := getPasswd(ctx, a.docker, *a.container)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to get the passwd file from container")

		return false
	}
	defer passwd.Close() //nolint:errcheck // response body of a finished read, nothing to flush

	user, err := osauth.LookupUserFromPasswd(username, passwd)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to lookup for the user on passwd file")

		return false
	}

	if user.Password == "" {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).Error("user passwd is empty, so the authentication via password is blocked")

		return false
	}

	shadow, err := getShadow(ctx, a.docker, *a.container)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to get the shadow file from container")

		return false
	}
	defer shadow.Close() //nolint:errcheck // response body of a finished read, nothing to flush

	if !osauth.AuthUserFromShadow(username, password, shadow) {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).Error("failed to authenticate the user on the device")

		return false
	}

	ctx.SetValue("user", user)

	log.WithFields(
		log.Fields{
			"container": *a.container,
			"username":  username,
		},
	).Info("using password authentication")

	return true
}

// PublicKey handles the server's SSH public key authentication when server is running in connector mode.
func (a *Authenticator) PublicKey(ctx gliderssh.Context, username string, key gliderssh.PublicKey) bool {
	passwd, err := getPasswd(ctx, a.docker, *a.container)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to get the passwd file from container")

		return false
	}
	defer passwd.Close() //nolint:errcheck // response body of a finished read, nothing to flush

	user, err := osauth.LookupUserFromPasswd(username, passwd)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to lookup for the user on passwd file")

		return false
	}

	if accountExpiredInContainer(ctx, a.docker, *a.container, username) {
		return false
	}

	type Signature struct {
		Username  string `json:"Username"`
		Namespace string `json:"Namespace"`
	}

	sig := &Signature{
		Username:  username,
		Namespace: *a.container,
	}

	sigBytes, err := json.Marshal(sig)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to marshal signature")

		return false
	}

	sigHash := sha256.Sum256(sigBytes)

	fingerprint := gossh.FingerprintLegacyMD5(key)
	res, err := a.api.AuthPublicKey(&models.PublicKeyAuthRequest{
		Fingerprint: fingerprint,
		Data:        string(sigBytes),
	}, a.authData.Token)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).WithError(err).Error("failed to authenticate the user via public key")

		return false
	}

	digest, err := base64.StdEncoding.DecodeString(res.Signature)
	if err != nil {
		if err != nil {
			log.WithFields(
				log.Fields{
					"container":   *a.container,
					"username":    username,
					"fingerprint": fingerprint,
				},
			).WithError(err).Error("failed to decode the signature")

			return false
		}

		return false
	}

	cryptoKey, ok := key.(gossh.CryptoPublicKey)
	if !ok {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).Error("failed to get the crypto public key")

		return false
	}

	pubCrypto := cryptoKey.CryptoPublicKey()

	pubKey, ok := pubCrypto.(*rsa.PublicKey)
	if !ok {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).Error("failed to convert the crypto public key")

		return false
	}

	if err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sigHash[:], digest); err != nil {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).WithError(err).Error("failed to verify the signature")

		return false
	}

	ctx.SetValue("user", user)

	log.WithFields(
		log.Fields{
			"container":   *a.container,
			"username":    username,
			"fingerprint": fingerprint,
		},
	).Info("using public key authentication")

	return true
}
