//go:build !freebsd

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
	"github.com/shellhub-io/shellhub/agent/pkg/yescrypt"
	"github.com/sirupsen/logrus"
)

var (
	// DefaultShadowFilename is where the hashed passwords are read from.
	DefaultShadowFilename = "/etc/shadow"
	// DefaultPasswdFilename is where the account records are read from.
	DefaultPasswdFilename = "/etc/passwd"
)

// DefaultBackend is the [Backend] used by the package-level helpers. Tests replace it to
// avoid reading the host's real account files.
var DefaultBackend Backend

type backend struct{}

func (b *backend) AuthUser(username, password string) bool {
	file, err := os.Open(DefaultShadowFilename)
	if err != nil {
		return false
	}
	defer file.Close() //nolint:errcheck

	return AuthUserFromShadow(username, password, file)
}

func (b *backend) LookupUser(username string) (*User, error) {
	file, err := os.Open(DefaultPasswdFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck

	return LookupUserFromPasswd(username, file)
}

func init() {
	DefaultBackend = &backend{}
}

type shadowEntry struct {
	Username    string
	Password    string
	Lastchanged int
	Minimum     int
	Maximum     int
	Warn        int
	Inactive    int
	Expire      int
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

// LookupUserFromPasswd try to find a [PasswordEntry] for a username from a passwd file.
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

		return nil, ErrUserNotFound
	}

	return &user, nil
}

// VerifyPasswordHash checks if the password match with the hash.
func VerifyPasswordHash(hash, password string) bool {
	if hash == "" {
		if PermitEmptyPasswords() {
			logrus.Warn("User logged in with empty password")

			return true
		}

		logrus.Error("User cannot login with empty password")

		return false
	}

	if hash == "!" || hash == "*" {
		logrus.Error("User cannot login with password")

		return false
	}

	if strings.HasPrefix(hash, "!") {
		logrus.Error("Password is locked")

		return false
	}

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

	for lineno := 1; ; lineno++ {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}

		if len(line) == 0 || strings.HasPrefix(string(line), "#") {
			continue
		}

		entry, err := parseShadowLine(string(line))
		if err != nil {
			return nil, fmt.Errorf("shadow line %d: %w", lineno, err)
		}

		entries[entry.Username] = entry
	}

	return entries, nil //nolint:nilerr
}

func parseShadowLine(line string) (shadowEntry, error) {
	result := shadowEntry{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 9 {
		return result, fmt.Errorf("wrong number of fields: %d != 9", len(parts))
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
	if value == "" {
		return 0
	}

	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}

	return number
}

func singleUser() *User {
	var uid, gid uint32
	var username, name, homeDir, shell string
	u, err := user.Current()
	uid, _ = parseUint32(os.Getenv("UID"))
	homeDir = os.Getenv("HOME")
	shell = os.Getenv("SHELL")
	if err == nil {
		uid, _ = parseUint32(u.Uid)
		gid, _ = parseUint32(u.Gid)
		username = u.Username
		name = u.Name
		homeDir = u.HomeDir
	}

	return &User{
		UID:      uid,
		GID:      gid,
		Username: username,
		Name:     name,
		HomeDir:  homeDir,
		Shell:    shell,
	}
}

func parsePasswdReader(r io.Reader) (map[string]User, error) {
	lines := bufio.NewReader(r)
	entries := make(map[string]User)
	for lineno := 1; ; lineno++ {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}

		if len(line) == 0 || strings.HasPrefix(string(line), "#") {
			continue
		}

		entry, err := parsePasswdLine(string(line))
		if err != nil {
			return nil, fmt.Errorf("passwd line %d: %w", lineno, err)
		}

		entries[entry.Username] = entry
	}

	return entries, nil //nolint:nilerr
}

func parsePasswdLine(line string) (User, error) {
	result := User{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 7 {
		return result, fmt.Errorf("wrong number of fields: %d != 7", len(parts))
	}
	result.Username = strings.TrimSpace(parts[0])
	result.Password = strings.TrimSpace(parts[1])

	uid, err := parseUint32(parts[2])
	if err != nil {
		return result, errMalformedUID
	}
	result.UID = uid

	gid, err := parseUint32(parts[3])
	if err != nil {
		return result, errMalformedGID
	}
	result.GID = gid

	result.Name = strings.TrimSpace(parts[4])
	result.HomeDir = strings.TrimSpace(parts[5])
	result.Shell = strings.TrimSpace(parts[6])

	return result, nil
}
