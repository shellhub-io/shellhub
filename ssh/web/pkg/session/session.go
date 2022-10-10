package session

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/cache"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
)

type Session struct {
	Token       string
	Device      string
	Username    string
	Password    string
	Fingerprint string
	Signature   string
}

func NewSession(ctx context.Context, device, username, password, fingerprint, signature string) (*Session, error) {
	key := magickey.GetRerefence()

	token, err := token.NewToken(uuid.Generate(), key)
	if err != nil {
		return nil, errors.New("failed to generate the session's token")
	}

	if password != "" {
		signed, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, []byte(password), nil)
		if err != nil {
			return nil, errors.New("failed to sign the session's password")
		}

		password = hex.EncodeToString(signed)
	}

	cached, err := cache.Token(ctx, token, struct {
		Device      string
		Username    string
		Password    string // TODO: encrypt this using the magic key.
		Fingerprint string
		Signature   string
	}{
		Device:      device,
		Username:    username,
		Password:    password,
		Fingerprint: fingerprint,
		Signature:   signature,
	})
	if err != nil {
		return nil, errors.New("failed to cache the session's token")
	}

	return &Session{ // nolint: exhaustruct
		Token:    cached.Token,
		Device:   cached.Device,
		Username: cached.Username,
		Password: cached.Password,
	}, nil
}

func Restore(ctx context.Context, jwt string) (*Session, error) {
	key := magickey.GetRerefence()

	token, err := token.Parse(jwt)
	if err != nil {
		return nil, errors.New("invalid session's token")
	}

	cached, err := cache.Restore(ctx, token)
	if err != nil {
		return nil, errors.New("failed to get credentials to login")
	}

	if cached.Password != "" {
		decoded, err := hex.DecodeString(cached.Password)
		if err != nil {
			return nil, errors.New("failed to decode the session's password")
		}

		decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, decoded, nil)
		if err != nil {
			return nil, errors.New("failed to decrypt the session's password")
		}

		cached.Password = string(decrypted)
	}

	return &Session{
		Token:       cached.Token,
		Device:      cached.Device,
		Username:    cached.Username,
		Password:    cached.Password,
		Fingerprint: cached.Fingerprint,
		Signature:   cached.Signature,
	}, nil
}
