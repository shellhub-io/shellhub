package web

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/cache"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
)

type Input struct {
	Device      string
	Username    string
	Password    string
	Fingerprint string
	Signature   string
}

type Output struct {
	Token string
}

type Session struct {
	Token       string
	Device      string
	Username    string
	Password    string
	Fingerprint string
	Signature   string
}

// CreateSession creates a new web session.
func CreateSession(ctx context.Context, data *Input) (*Session, error) {
	if data == nil {
		return nil, errors.New("failed to get the session's data")
	}

	key := magickey.GetRerefence()

	token, err := token.NewToken(key)
	if err != nil {
		return nil, errors.New("failed to generate the session's token")
	}

	if data.Password != "" {
		signed, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, []byte(data.Password), nil)
		if err != nil {
			return nil, errors.New("failed to sign the session's password")
		}

		data.Password = hex.EncodeToString(signed)
	}

	cached, err := cache.Save(ctx, token, &cache.Data{
		Device:      data.Device,
		Username:    data.Username,
		Password:    data.Password,
		Fingerprint: data.Fingerprint,
		Signature:   data.Signature,
	})
	if err != nil {
		return nil, errors.New("failed to cache the session's token")
	}

	return &Session{
		Token:       cached.Token,
		Device:      data.Device,
		Username:    data.Username,
		Password:    data.Password,
		Fingerprint: data.Fingerprint,
		Signature:   data.Signature,
	}, nil
}

// RestoreSession restores a web session.
func RestoreSession(ctx context.Context, data *Output) (*Session, error) {
	if data == nil {
		return nil, errors.New("failed to get the session's token")
	}

	key := magickey.GetRerefence()

	token, err := token.Parse(data.Token)
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
		Token:       data.Token,
		Device:      cached.Device,
		Username:    cached.Username,
		Password:    cached.Password,
		Fingerprint: cached.Fingerprint,
		Signature:   cached.Signature,
	}, nil
}
