package osauth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var DefaultPasswdFilename = "/etc/passwd"

type User struct {
	UID      uint32
	GID      uint32
	Username string
	Password string
	Name     string
	HomeDir  string
	Shell    string
}

func LookupUser(username string) *User {
	passwdFile, err := os.Open(DefaultPasswdFilename)
	if err != nil {
		logrus.Errorf("Could not open %s", DefaultPasswdFilename)
		return nil
	}
	defer passwdFile.Close()

	entries, err := parsePasswdReader(passwdFile)
	if err != nil {
		logrus.Printf("Could not parse passwdfile %v", err)
		return nil
	}

	user, found := entries[username]
	if !found {
		logrus.Error("User not found")
		return nil
	}

	return &user
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
	return entries, nil
}

func parsePasswdLine(line string) (User, error) {
	result := User{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 7 {
		return result, fmt.Errorf("Passwd line had wrong number of parts %d != 7", len(parts))
	}
	result.Username = strings.TrimSpace(parts[0])
	result.Password = strings.TrimSpace(parts[1])

	uid, err := strconv.Atoi(parts[2])
	if err != nil {
		return result, fmt.Errorf("Passwd line had badly formatted uid %s", parts[2])
	}
	result.UID = uint32(uid)

	gid, err := strconv.Atoi(parts[3])
	if err != nil {
		return result, fmt.Errorf("Passwd line had badly formatted gid %s", parts[3])
	}
	result.GID = uint32(gid)

	result.Name = strings.TrimSpace(parts[4])
	result.HomeDir = strings.TrimSpace(parts[5])
	result.Shell = strings.TrimSpace(parts[6])
	return result, nil
}
