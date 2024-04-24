package osauth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"    // GehirnInc/crypt uses blank imports for crypto subpackages
	_ "github.com/GehirnInc/crypt/sha256_crypt" // GehirnInc/crypt uses blank imports for crypto subpackages
	_ "github.com/GehirnInc/crypt/sha512_crypt" // GehirnInc/crypt uses blank imports for crypto subpackages
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/yescrypt"
	"github.com/sirupsen/logrus"
)

//go:generate mockery --name=Backend --filename=backend.go
type Backend interface {
	AuthUser(username, password string) bool
	LookupUser(username string) (*User, error)
}

var DefaultBackend Backend

func init() {
	DefaultBackend = &backend{}
}

var (
	// Default file path for shadow file.
	DefaultShadowFilename = "/etc/shadow"
	// Default file path for passwd file.
	DefaultPasswdFilename = "/etc/passwd"
)

// This struct represents an entry from the `/etc/shadow` file.
type shadowEntry struct {
	// The login name of the account (same as in [PasswdEntry]).
	Username string
	// The hashed password of the account (same as in [PasswdEntry]).
	Password string
	// The number of days since January 1, 1970 (epoch) that the password was last changed.
	Lastchanged int
	// The minimum number of days required between password changes i.e. the number of days left before the user is
	// allowed to change his/her password.
	Minimum int
	// The maximum number of days the password is valid (after that user is forced to change his/her password).
	Maximum int
	// The number of days before password is to expire that user is warned that his/her password must be changed.
	Warn int
	// The number of days after password expires that account is disabled.
	Inactive int
	// Days since Jan 1, 1970 that account is disabled i.e. an absolute date specifying when the login may no longer be
	// used.
	Expire int
}

// AuthUser attempts to authenticate username and password from [DefaultPasswdFilename].
func AuthUser(username, password string) bool {
	return DefaultBackend.AuthUser(username, password)
}

// LookupUser try to find a [PasswordEntry] for a username from a [DefaultPasswdFilename].
func LookupUser(username string) (*User, error) {
	return DefaultBackend.LookupUser(username)
}

// AuthUserFromShadow attempts to authenticate username and password from file.
func AuthUserFromShadow(username, password string, shadow io.Reader) bool {
	entries, err := parseShadowReader(shadow)
	if err != nil {
		logrus.WithError(err).Debug("Error parsing shadow file")

		return false
	}

	entry, ok := entries[username]
	if !ok {
		logrus.WithFields(logrus.Fields{
			"username": username,
		}).Error("User not found")

		return false
	}

	return VerifyPasswordHash(entry.Password, password)
}

// Lookup try to find a [PasswordEntry] for a username from a passwd file.
func LookupUserFromPasswd(username string, passwd io.Reader) (*User, error) {
	if os.Geteuid() != 0 {
		return singleUser(), nil
	}

	entries, err := parsePasswdReader(passwd)
	if err != nil {
		logrus.WithError(err).Error("Error parsing passwd file")

		return nil, err
	}

	user, found := entries[username]
	if !found {
		logrus.WithFields(logrus.Fields{
			"username": username,
		}).Error("User not found in passwd file")

		return nil, err
	}

	return &user, nil
}

// VerifyPasswordHash checks if the password match with the hash.
func VerifyPasswordHash(hash, password string) bool {
	if hash == "" {
		logrus.Error("Password entry is empty")

		return false
	}

	// NOTE: If hash algorithm is yescrypt, we verify by ourselves, otherwise let's try crypt package.
	if strings.HasPrefix(hash, "$y$") {
		return yescrypt.Verify(password, hash)
	}

	if ok := crypt.IsHashSupported(hash); !ok {
		logrus.Error("The crypto algorithm is not supported")

		return false
	}

	crypt := crypt.NewFromHash(hash)
	if crypt == nil {
		logrus.Error("Could not detect password crypto algorithm from shadow entry")

		return false
	}

	if err := crypt.Verify(hash, []byte(password)); err != nil {
		logrus.WithError(err).Debug("Error verifying password hash")

		return false
	}

	return true
}

func parseShadowReader(r io.Reader) (map[string]shadowEntry, error) {
	lines := bufio.NewReader(r)
	entries := make(map[string]shadowEntry)

	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}

		if len(line) == 0 || strings.HasPrefix(string(line), "#") {
			continue
		}

		entry, err := parseShadowLine(string(line))
		if err != nil {
			return nil, err
		}

		entries[entry.Username] = entry
	}

	return entries, nil //nolint:nilerr
}

func parseShadowLine(line string) (shadowEntry, error) {
	result := shadowEntry{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 9 {
		return result, fmt.Errorf("shadow line had wrong number of parts %d != 9", len(parts))
	}

	result.Username = strings.TrimSpace(parts[0])
	result.Password = strings.TrimSpace(parts[1])

	result.Lastchanged = parseIntString(parts[2])
	result.Minimum = parseIntString(parts[3])
	result.Maximum = parseIntString(parts[4])
	result.Warn = parseIntString(parts[5])
	result.Inactive = parseIntString(parts[6])
	result.Expire = parseIntString(parts[7])

	return result, nil
}

func parseIntString(value string) int {
	if value != "" {
		return 0
	}

	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}

	return number
}

func singleUser() *User {
	var uid, gid int
	var username, name, homeDir, shell string
	u, err := user.Current()
	uid, _ = strconv.Atoi(os.Getenv("UID"))
	homeDir = os.Getenv("HOME")
	shell = os.Getenv("SHELL")
	if err == nil {
		uid, _ = strconv.Atoi(u.Uid)
		gid, _ = strconv.Atoi(u.Gid)
		username = u.Username
		name = u.Name
		homeDir = u.HomeDir
	}

	return &User{
		UID:      uint32(uid),
		GID:      uint32(gid),
		Username: username,
		Name:     name,
		HomeDir:  homeDir,
		Shell:    shell,
	}
}

func parsePasswdReader(r io.Reader) (map[string]User, error) {
	lines := bufio.NewReader(r)
	entries := make(map[string]User)
	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}

		if len(line) == 0 || strings.HasPrefix(string(line), "#") {
			continue
		}

		entry, err := parsePasswdLine(string(line))
		if err != nil {
			return nil, err
		}

		entries[entry.Username] = entry
	}

	return entries, nil //nolint:nilerr
}

func parsePasswdLine(line string) (User, error) {
	result := User{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 7 {
		return result, fmt.Errorf("passwd line had wrong number of parts %d != 7", len(parts))
	}
	result.Username = strings.TrimSpace(parts[0])
	result.Password = strings.TrimSpace(parts[1])

	uid, err := strconv.Atoi(parts[2])
	if err != nil {
		return result, fmt.Errorf("passwd line had badly formatted uid %s", parts[2])
	}
	result.UID = uint32(uid)

	gid, err := strconv.Atoi(parts[3])
	if err != nil {
		return result, fmt.Errorf("passwd line had badly formatted gid %s", parts[3])
	}
	result.GID = uint32(gid)

	result.Name = strings.TrimSpace(parts[4])
	result.HomeDir = strings.TrimSpace(parts[5])
	result.Shell = strings.TrimSpace(parts[6])

	return result, nil
}
