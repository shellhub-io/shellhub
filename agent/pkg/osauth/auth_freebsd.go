//go:build freebsd
// +build freebsd

package osauth

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"    // GehirnInc/crypt uses blank imports for crypto subpackages
	_ "github.com/GehirnInc/crypt/sha256_crypt" // GehirnInc/crypt uses blank imports for crypto subpackages
	_ "github.com/GehirnInc/crypt/sha512_crypt" // GehirnInc/crypt uses blank imports for crypto subpackages
	"github.com/shellhub-io/shellhub/pkg/clock"
	log "github.com/sirupsen/logrus"
)

var DefaultMasterPasswdFilename = "/etc/master.passwd"

var (
	errMalformedChange = errors.New("malformed change field")
	errMalformedExpire = errors.New("malformed expire field")
)

var DefaultBackend Backend

type backend struct{}

type masterPasswdEntry struct {
	User
	Change int64
	Expire int64
}

func (e masterPasswdEntry) checkAccountExpiry(now int64) error {
	if e.Expire != 0 && now >= e.Expire {
		return errAccountExpired
	}

	return nil
}

func (e masterPasswdEntry) checkPasswordLogin(now int64) error {
	if err := e.checkAccountExpiry(now); err != nil {
		return err
	}

	if e.Change != 0 && now >= e.Change {
		return errPasswordExpired
	}

	return nil
}

func (b *backend) AuthUser(username, password string) bool {
	file, err := os.Open(DefaultMasterPasswdFilename)
	if err != nil {
		return false
	}

	return AuthUserFromShadow(username, password, file)
}

func (b *backend) AccountExpired(username string) bool {
	file, err := os.Open(DefaultMasterPasswdFilename)
	if err != nil {
		log.WithError(err).Error("Error opening master.passwd file")

		return true
	}
	defer file.Close() //nolint:errcheck

	return AccountExpiredFromShadow(username, file)
}

func (b *backend) LookupUser(username string) (*User, error) {
	file, err := os.Open(DefaultMasterPasswdFilename)
	if err != nil {
		return nil, err
	}

	return LookupUserFromPasswd(username, file)
}

func init() {
	DefaultBackend = &backend{}
}

// AuthUser attempts to authenticate username and password from the [DefaultMasterPasswdFilename].
func AuthUser(username, password string) bool {
	return DefaultBackend.AuthUser(username, password)
}

// LookupUser try to find a [PasswordEntry] for a username from the [DefaultMasterPasswdFilename].
func LookupUser(username string) (*User, error) {
	return DefaultBackend.LookupUser(username)
}

// AccountExpired reports whether the account's entry in [DefaultMasterPasswdFilename] has expired,
// returning true when the file cannot be read.
func AccountExpired(username string) bool {
	return DefaultBackend.AccountExpired(username)
}

// AuthUserFromShadow attempts to authenticate username and password from file. It refuses an
// account past its expire time, or whose password is past its change time, even when the
// password matches, because the agent cannot run the change a local login would.
func AuthUserFromShadow(username, password string, shadow io.Reader) bool {
	entries, err := parseMasterPasswdReader(shadow)
	if err != nil {
		log.WithError(err).Error("Error parsing passwd file")

		return false
	}

	user, found := entries[username]
	if !found {
		log.WithFields(log.Fields{
			"username": username,
		}).Error("User not found in passwd file")

		return false
	}

	if !VerifyPasswordHash(user.Password, password) {
		return false
	}

	if err := user.checkPasswordLogin(clock.Now().Unix()); err != nil {
		log.WithFields(log.Fields{
			"username": username,
		}).WithError(err).Error("Refusing login")

		return false
	}

	return true
}

// AccountExpiredFromShadow reports whether the account's entry in masterPasswd has expired. It
// returns true when masterPasswd cannot be parsed, and false when the account has no entry.
func AccountExpiredFromShadow(username string, masterPasswd io.Reader) bool {
	entries, err := parseMasterPasswdReader(masterPasswd)
	if err != nil {
		log.WithError(err).Error("Error parsing passwd file")

		return true
	}

	entry, found := entries[username]
	if !found {
		return false
	}

	if err := entry.checkAccountExpiry(clock.Now().Unix()); err != nil {
		log.WithFields(log.Fields{
			"username": username,
		}).WithError(err).Error("Refusing login")

		return true
	}

	return false
}

// Lookup try to find a [PasswordEntry] for a username from a passwd file.
func LookupUserFromPasswd(username string, passwd io.Reader) (*User, error) {
	entries, err := parseMasterPasswdReader(passwd)
	if err != nil {
		log.WithError(err).Error("Error parsing passwd file")

		return nil, err
	}

	user, found := entries[username]
	if !found {
		log.WithFields(log.Fields{
			"username": username,
		}).Error("User not found in passwd file")

		return nil, ErrUserNotFound
	}

	return &user.User, nil
}

func parseMasterPasswdReader(r io.Reader) (map[string]masterPasswdEntry, error) {
	lines := bufio.NewReader(r)
	entries := make(map[string]masterPasswdEntry)
	for lineno := 1; ; lineno++ {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}

		if len(line) == 0 || strings.HasPrefix(string(line), "#") {
			continue
		}

		entry, err := parseMasterPasswdLine(string(line))
		if errors.Is(err, errMalformedChange) || errors.Is(err, errMalformedExpire) {
			log.WithError(err).Warnf("Skipping master.passwd line %d", lineno)

			continue
		}

		if err != nil {
			return nil, fmt.Errorf("master.passwd line %d: %w", lineno, err)
		}

		entries[entry.Username] = entry
	}

	return entries, nil //nolint:nilerr
}

func parseMasterPasswdLine(line string) (masterPasswdEntry, error) {
	result := masterPasswdEntry{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 10 {
		return result, fmt.Errorf("wrong number of fields: %d != 10", len(parts))
	}
	result.Username = strings.TrimSpace(parts[0])
	result.Password = strings.TrimSpace(parts[1])

	uid, err := strconv.Atoi(parts[2])
	if err != nil {
		return result, errMalformedUID
	}
	result.UID = uint32(uid)

	gid, err := strconv.Atoi(parts[3])
	if err != nil {
		return result, errMalformedGID
	}
	result.GID = uint32(gid)

	change, err := parseEpochSeconds(parts[5])
	if err != nil {
		return result, errMalformedChange
	}
	result.Change = change

	expire, err := parseEpochSeconds(parts[6])
	if err != nil {
		return result, errMalformedExpire
	}
	result.Expire = expire

	result.HomeDir = strings.TrimSpace(parts[8])
	result.Shell = strings.TrimSpace(parts[9])

	return result, nil
}

func parseEpochSeconds(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	return strconv.ParseInt(value, 10, 64)
}

// VerifyPasswordHash checks if the password match with the hash.
func VerifyPasswordHash(hash, password string) bool {
	if hash == "" {
		if PermitEmptyPasswords() {
			log.Warn("User logged in with empty password")

			return true
		}

		log.Error("User cannot login with empty password")

		return false
	}

	if password == "" && (hash == "*LOCKED*" || hash == "*") {
		log.Error("Password is locked")

		return false
	}

	if hash != "" && password == "" {
		log.Error("Password entry is empty")

		return false
	}

	if ok := crypt.IsHashSupported(hash); !ok {
		log.Error("The crypto algorithm is not supported")

		return false
	}

	crypt := crypt.NewFromHash(hash)
	if crypt == nil {
		log.Error("Could not detect password crypto algorithm from shadow entry")

		return false
	}

	if err := crypt.Verify(hash, []byte(password)); err != nil {
		log.WithError(err).Debug("Error verifying password hash")

		return false
	}

	return true
}
