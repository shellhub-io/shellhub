// Package cache provides a interface to store and retrieve session's data from a cache.
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
)

// instance is a singleton cache instance.
var instance cache.Cache

// TTL is the time to live of the token in the cache.
const TTL = time.Second * 30

// Data is the data set to be saved in the cache.
type Data struct {
	// Device is the device's name.
	Device string
	// Username is the username of the user to login.
	Username string
	// Password is the password of the user to login.
	// Password is should be empty if the user is using a public key.
	Password string
	// Fingerprint is the fingerprint of the public key.
	// Fingerprint is should be empty if the user is using a password.
	Fingerprint string
	// Signature is the signature of the public key.
	// Signature is should be empty if the user is using a password.
	Signature string
}

type Token struct {
	// ID is the token's identifier.
	ID string
	// Token is the JWT token.
	Token string
}

// ConnectRedis connects to redis to be used as cache system.
func ConnectRedis(uri string) error {
	if instance == nil {
		var err error
		instance, err = cache.NewRedisCache(uri)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func getConnection() (cache.Cache, error) { //nolint: ireturn
	if instance == nil {
		return nil, errors.New("cache was not connected")
	}

	return instance, nil
}

// Save saves a data set for TTL time using token as identifier.
func Save(ctx context.Context, token *token.Token, data *Data) (*Token, error) {
	connection, err := getConnection()
	if err != nil {
		return nil, err
	}

	if err := connection.Set(ctx, token.ID, data, TTL); err != nil {
		return nil, err
	}

	return &Token{
		ID:    token.ID,
		Token: token.Data,
	}, nil
}

// Restore restores a data set using token as identifier.
func Restore(ctx context.Context, token *token.Token) (*Data, error) {
	connection, err := getConnection()
	if err != nil {
		return nil, err
	}

	var value struct {
		Device      string
		Username    string
		Password    string
		Fingerprint string
		Signature   string
	}

	if err := connection.Get(ctx, token.ID, &value); err != nil {
		return nil, err
	}

	return &Data{
		Device:      value.Device,
		Username:    value.Username,
		Password:    value.Password,
		Fingerprint: value.Fingerprint,
		Signature:   value.Signature,
	}, nil
}
