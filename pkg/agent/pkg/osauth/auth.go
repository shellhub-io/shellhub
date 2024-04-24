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
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/yescrypt"
	"github.com/sirupsen/logrus"
)

func AuthUser(username, password string) bool {
	shadow, err := os.Open(DefaultShadowFilename)
	if err != nil {
		logrus.WithError(err).Error("Could not open /etc/shadow")

		return false
	}
	defer shadow.Close()

	if ok := AuthUserFromShadow(username, password, shadow); !ok {
		logrus.WithFields(logrus.Fields{
			"username": username,
		}).Debug("Failed to authenticate user from shadow file")

		return false
	}

	return true
}

// AuthUserFromShadow checks if the given username and password are valid for the given shadow file.
func AuthUserFromShadow(username string, password string, shadow io.Reader) bool {
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

func VerifyPasswordHash(hash, password string) bool {
	if hash == "" {
		logrus.Error("Password entry is empty")

		return false
	}

	// If hash algorithm is yescrypt verify by ourselves, otherwise let's try crypt package
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

// ErrUserNotFound is returned when the user is not found in the passwd file.
var ErrUserNotFound = errors.New("user not found")

// LookupUserFromPasswd reads the passwd file from the given reader and returns the user, if found.
// TODO: Use this function inside the LookupUser.
func LookupUserFromPasswd(username string, passwd io.Reader) (*User, error) {
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

		return nil, ErrUserNotFound
	}

	return &user, nil
}

func LookupUser(username string) *User {
	if os.Geteuid() != 0 {
		return singleUser()
	}

	passwd, err := os.Open(DefaultPasswdFilename)
	if err != nil {
		logrus.Errorf("Could not open %s", DefaultPasswdFilename)

		return nil
	}
	defer passwd.Close()

	user, err := LookupUserFromPasswd(username, passwd)
	if err != nil {
		return nil
	}

	return user
}

var DefaultShadowFilename = "/etc/shadow"

type ShadowEntry struct {
	Username    string // Login name
	Password    string // Hashed password
	Lastchanged int    // Days since Jan 1, 1970 that password was last changed
	Minimum     int    // The minimum number of days required between password changes i.e. the number of days left before the user is allowed to change his/her password
	Maximum     int    // The maximum number of days the password is valid (after that user is forced to change his/her password)
	Warn        int    // The number of days before password is to expire that user is warned that his/her password must be changed
	Inactive    int    // The number of days after password expires that account is disabled
	Expire      int    // Days since Jan 1, 1970 that account is disabled i.e. an absolute date specifying when the login may no longer be used.
}

func parseShadowReader(r io.Reader) (map[string]ShadowEntry, error) {
	lines := bufio.NewReader(r)
	entries := make(map[string]ShadowEntry)

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

func parseShadowLine(line string) (ShadowEntry, error) {
	result := ShadowEntry{}
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
