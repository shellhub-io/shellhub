//go:build freebsd
// +build freebsd

package osauth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"    // GehirnInc/crypt uses blank imports for crypto subpackages
	_ "github.com/GehirnInc/crypt/sha256_crypt" // GehirnInc/crypt uses blank imports for crypto subpackages
	_ "github.com/GehirnInc/crypt/sha512_crypt" // GehirnInc/crypt uses blank imports for crypto subpackages
	log "github.com/sirupsen/logrus"
)

var DefaultMasterPasswdFilename = "/etc/master.passwd"

var DefaultBackend Backend

type backend struct{}

func (b *backend) AuthUser(username, password string) bool {
	file, err := os.Open(DefaultMasterPasswdFilename)
	if err != nil {
		return false
	}

	return AuthUserFromShadow(username, password, file)
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

// AuthUserFromShadow attempts to authenticate username and password from file.
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

	return VerifyPasswordHash(user.Password, password)
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

		return nil, err
	}

	return &user, nil
}

func parseMasterPasswdReader(r io.Reader) (map[string]User, error) {
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

		entry, err := parseMasterPasswdLine(string(line))
		if err != nil {
			return nil, err
		}

		entries[entry.Username] = entry
	}

	return entries, nil //nolint:nilerr
}

func parseMasterPasswdLine(line string) (User, error) {
	result := User{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 10 {
		return result, fmt.Errorf("passwd line had wrong number of parts %d != 10", len(parts))
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

	// result.Name = strings.TrimSpace(parts[4])

	result.HomeDir = strings.TrimSpace(parts[8])
	result.Shell = strings.TrimSpace(parts[9])

	return result, nil
}

// VerifyPasswordHash checks if the password match with the hash.
func VerifyPasswordHash(hash, password string) bool {
	if hash == "" {
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
