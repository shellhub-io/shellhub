package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const fakeLockoutDuration = 60

type fakeLockoutCache struct {
	cache.Cache

	mu       sync.Mutex
	now      int64
	held     map[string]string
	attempts map[string]int
	lockouts map[string]int64
}

func newFakeLockoutCache(now time.Time) *fakeLockoutCache {
	return &fakeLockoutCache{
		Cache:    cache.NewNullCache(),
		now:      now.Unix(),
		held:     map[string]string{},
		attempts: map[string]int{},
		lockouts: map[string]int64{},
	}
}

func (c *fakeLockoutCache) SetNX(_ context.Context, key string, value any, _ time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.held[key]; ok {
		return false, nil
	}

	owner, _ := value.(string)
	c.held[key] = owner

	return true, nil
}

func (c *fakeLockoutCache) CompareAndDelete(_ context.Context, key, value string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.held[key] != value {
		return false, nil
	}

	delete(c.held, key)

	return true, nil
}

func (c *fakeLockoutCache) HasAccountLockout(_ context.Context, source, id string) (int64, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := source + ":" + id
	if c.lockouts[key] > c.now {
		return c.lockouts[key], c.attempts[key], nil
	}

	return 0, c.attempts[key], nil
}

func (c *fakeLockoutCache) StoreLoginAttempt(_ context.Context, source, id string) (int64, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := source + ":" + id
	c.attempts[key]++
	if c.attempts[key] < 3 {
		return 0, c.attempts[key], nil
	}

	c.lockouts[key] = c.now + fakeLockoutDuration

	return c.lockouts[key], c.attempts[key], nil
}

func (c *fakeLockoutCache) ResetLoginAttempts(_ context.Context, source, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.attempts, source+":"+id)
	delete(c.lockouts, source+":"+id)

	return nil
}

func (c *fakeLockoutCache) releasedEveryGuard() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.held) == 0
}

func guessConcurrently(t *testing.T, hashMock *hashmock.MockHasher, digest string, n int, guess func() error) []error {
	t.Helper()

	checking := make(chan struct{})
	finish := make(chan struct{})
	hashMock.On("CompareWith", "wrong", digest).
		Run(func(mock.Arguments) {
			close(checking)
			<-finish
		}).
		Return(false).
		Once()

	errs := make([]error, n)

	var admitted sync.WaitGroup
	admitted.Go(func() { errs[0] = guess() })
	<-checking

	var refused sync.WaitGroup
	for i := 1; i < n; i++ {
		refused.Go(func() { errs[i] = guess() })
	}

	refused.Wait()
	close(finish)
	admitted.Wait()

	return errs
}

func requireLockedOutUntil(t *testing.T, err error, until int64) {
	t.Helper()

	require.ErrorIs(t, err, ErrFactorLockedOut)

	var e errors.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, ErrDataLockout{Until: until}, e.Data)
}

func TestWebReauthVerifyLimitsPasswordGuesses(t *testing.T) {
	ctx := context.Background()

	const (
		tenantID    = "00000000-0000-4000-0000-000000000000"
		userID      = "00000000-0000-0000-0000-00000000000a"
		fingerprint = "SHA256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		digest      = "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"
		password    = "correct-horse"
	)

	user := &models.User{ID: userID, Password: models.UserPassword{Hash: digest}}
	lockoutEnd := now.Unix() + fakeLockoutDuration

	verify := func(pw string) *requests.WebReauthVerify {
		return &requests.WebReauthVerify{TenantID: tenantID, UserID: userID, Password: pw, Fingerprint: fingerprint}
	}

	setup := func(t *testing.T) (*storemock.MockStore, *hashmock.MockHasher, *fakeLockoutCache, Service) {
		t.Helper()

		hashMock := hashmock.NewMockHasher(t)
		hash.Set(t, hashMock)

		storeMock := storemock.NewMockStore(t)
		storeMock.On("UserResolve", mock.Anything, store.UserIDResolver, userID).Return(user, nil).Maybe()
		storeMock.On("SSHIdentityResolve", mock.Anything, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
			Return(&models.SSHIdentity{PrincipalID: userID, Fingerprint: fingerprint}, nil).Maybe()
		storeMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
			Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Maybe()
		storeMock.On("SSHIdentityTouchReauth", mock.Anything, tenantID, fingerprint).Return(nil).Maybe()

		fake := newFakeLockoutCache(now)
		service := NewService(store.Store(storeMock), privateKey, publicKey, fake, WithClock(clock.NewFixed(now)))

		return storeMock, hashMock, fake, service
	}

	wrongGuesses := func(t *testing.T, hashMock *hashmock.MockHasher, service Service, n int) []error {
		t.Helper()

		hashMock.On("CompareWith", "wrong", digest).Return(false).Times(n)

		errs := make([]error, 0, n)
		for range n {
			_, err := service.WebReauthVerify(ctx, verify("wrong"))
			errs = append(errs, err)
		}

		return errs
	}

	t.Run("an unknown fingerprint answers the same whatever the password, checking none", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("UserResolve", mock.Anything, store.UserIDResolver, userID).Return(user, nil)
		storeMock.On("SSHIdentityResolve", mock.Anything, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
			Return(nil, store.ErrNoDocuments)
		hash.Set(t, hashmock.NewMockHasher(t))

		service := NewService(store.Store(storeMock), privateKey, publicKey, newFakeLockoutCache(now), WithClock(clock.NewFixed(now)))

		_, right := service.WebReauthVerify(ctx, verify(password))
		_, wrong := service.WebReauthVerify(ctx, verify("wrong"))

		require.ErrorIs(t, right, ErrSSHIdentityNotFound)
		assert.Equal(t, right.Error(), wrong.Error())
	})

	t.Run("the third wrong password starts a lockout", func(t *testing.T) {
		_, hashMock, fake, service := setup(t)

		errs := wrongGuesses(t, hashMock, service, 3)

		require.ErrorIs(t, errs[0], ErrUserPasswordNotMatch)
		require.ErrorIs(t, errs[1], ErrUserPasswordNotMatch)
		requireLockedOutUntil(t, errs[2], lockoutEnd)
		assert.True(t, fake.releasedEveryGuard())
	})

	t.Run("a password sent during the lockout is refused unchecked", func(t *testing.T) {
		_, hashMock, _, service := setup(t)
		wrongGuesses(t, hashMock, service, 3)

		_, err := service.WebReauthVerify(ctx, verify(password))

		requireLockedOutUntil(t, err, lockoutEnd)
	})

	t.Run("of concurrent guesses only one is checked, the rest refused unchecked", func(t *testing.T) {
		_, hashMock, fake, service := setup(t)

		errs := guessConcurrently(t, hashMock, digest, 8, func() error {
			_, err := service.WebReauthVerify(ctx, verify("wrong"))

			return err
		})

		require.ErrorIs(t, errs[0], ErrUserPasswordNotMatch)
		for _, err := range errs[1:] {
			requireLockedOutUntil(t, err, 0)
		}
		assert.True(t, fake.releasedEveryGuard())
	})

	t.Run("an empty password is refused without counting toward the lockout", func(t *testing.T) {
		_, hashMock, fake, service := setup(t)

		for range 5 {
			_, err := service.WebReauthVerify(ctx, verify(""))
			require.ErrorIs(t, err, ErrUserPasswordNotMatch)
		}

		errs := wrongGuesses(t, hashMock, service, 2)
		require.ErrorIs(t, errs[1], ErrUserPasswordNotMatch)
		assert.True(t, fake.releasedEveryGuard())
	})

	t.Run("a right password clears the count", func(t *testing.T) {
		_, hashMock, _, service := setup(t)
		wrongGuesses(t, hashMock, service, 2)

		hashMock.On("CompareWith", password, digest).Return(true).Once()
		_, err := service.WebReauthVerify(ctx, verify(password))
		require.NoError(t, err)

		errs := wrongGuesses(t, hashMock, service, 2)
		require.ErrorIs(t, errs[1], ErrUserPasswordNotMatch)
	})
}

func TestUpdateUserLimitsCurrentPasswordGuesses(t *testing.T) {
	ctx := context.Background()

	const (
		userID = "000000000000000000000000"
		digest = "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"
	)

	user := func() *models.User {
		return &models.User{
			ID:       userID,
			UserData: models.UserData{Name: "James Smith", Username: "james_smith", Email: "james@shellhub.test"},
			Password: models.UserPassword{Hash: digest},
		}
	}

	change := func(current string) *requests.UpdateUser {
		return &requests.UpdateUser{UserID: userID, Password: "new-secret", CurrentPassword: current}
	}

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)

	storeMock := storemock.NewMockStore(t)
	storeMock.On("UserResolve", ctx, store.UserIDResolver, userID).Return(user(), nil).Times(4)
	hashMock.On("CompareWith", "wrong", digest).Return(false).Times(3)

	service := NewService(store.Store(storeMock), privateKey, publicKey, newFakeLockoutCache(now), WithClock(clock.NewFixed(now)))

	require.ErrorIs(t, service.UpdateUser(ctx, change("wrong")), ErrUserPasswordNotMatch)
	require.ErrorIs(t, service.UpdateUser(ctx, change("wrong")), ErrUserPasswordNotMatch)
	requireLockedOutUntil(t, service.UpdateUser(ctx, change("wrong")), now.Unix()+fakeLockoutDuration)

	requireLockedOutUntil(t, service.UpdateUser(ctx, change("secret")), now.Unix()+fakeLockoutDuration)
}

func TestUpdateUserAdmitsOneConcurrentCurrentPasswordGuess(t *testing.T) {
	ctx := context.Background()

	const (
		userID  = "000000000000000000000000"
		digest  = "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"
		guesses = 8
	)

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)

	storeMock := storemock.NewMockStore(t)
	storeMock.On("UserResolve", ctx, store.UserIDResolver, userID).
		Return(func(context.Context, store.UserResolver, string, ...store.QueryOption) (*models.User, error) {
			return &models.User{ID: userID, Password: models.UserPassword{Hash: digest}}, nil
		}).
		Times(guesses)

	fake := newFakeLockoutCache(now)
	service := NewService(store.Store(storeMock), privateKey, publicKey, fake, WithClock(clock.NewFixed(now)))

	errs := guessConcurrently(t, hashMock, digest, guesses, func() error {
		return service.UpdateUser(ctx, &requests.UpdateUser{UserID: userID, Password: "new-secret", CurrentPassword: "wrong"})
	})

	require.ErrorIs(t, errs[0], ErrUserPasswordNotMatch)
	for _, err := range errs[1:] {
		requireLockedOutUntil(t, err, 0)
	}
	assert.True(t, fake.releasedEveryGuard())
}
