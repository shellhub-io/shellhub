package osauth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// DefaultGroupFilename is where the group memberships are read from. A Unix group file has the
// same colon-separated layout on every platform the agent runs on, so the parsing below is shared
// rather than reimplemented per GOOS.
var DefaultGroupFilename = "/etc/group"

func (b *backend) ListGroups(username string) ([]uint32, error) {
	user, err := b.LookupUser(username)
	if err != nil {
		return nil, err
	}

	groups := []uint32{user.GID}

	groupFile, err := os.Open(DefaultGroupFilename)
	if err != nil {
		return nil, err
	}
	defer groupFile.Close() //nolint:errcheck

	secondaryGroups, err := ListGroupsFromFile(username, groupFile)
	if err != nil {
		return nil, err
	}

	for _, gid := range secondaryGroups {
		if gid != user.GID {
			groups = append(groups, gid)
		}
	}

	return groups, nil
}

// ListGroups returns a list of group IDs that the user belongs to.
func ListGroups(username string) ([]uint32, error) {
	return DefaultBackend.ListGroups(username)
}

// ListGroupsFromFile returns the GIDs of every group in group that lists username as a member.
// The primary group recorded in passwd is not among them.
func ListGroupsFromFile(username string, group io.Reader) ([]uint32, error) {
	groups, err := parseGroupReader(group)
	if err != nil {
		logrus.WithError(err).Error("Error parsing group file")

		return nil, err
	}

	var userGroups []uint32
	for _, g := range groups {
		if slices.Contains(g.Members, username) {
			userGroups = append(userGroups, g.GID)
		}
	}

	return userGroups, nil
}

// Group is one entry of a Unix group file.
type Group struct {
	Name     string   // The name of the group.
	Password string   // The password of the group.
	GID      uint32   // The group ID of the group.
	Members  []string // The list of members in the group.
}

func parseGroupLine(line string) (Group, error) {
	result := Group{}
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 4 {
		return result, fmt.Errorf("wrong number of fields: %d != 4", len(parts))
	}
	result.Name = strings.TrimSpace(parts[0])
	result.Password = strings.TrimSpace(parts[1])

	gid, err := parseUint32(parts[2])
	if err != nil {
		return result, errMalformedGID
	}
	result.GID = gid

	members := strings.TrimSpace(parts[3])
	if members != "" {
		result.Members = strings.Split(members, ",")
	} else {
		result.Members = []string{}
	}

	return result, nil
}

func parseGroupReader(r io.Reader) (map[string]Group, error) {
	lines := bufio.NewReader(r)
	entries := make(map[string]Group)
	for lineno := 1; ; lineno++ {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}

		if len(line) == 0 || strings.HasPrefix(string(line), "#") {
			continue
		}

		entry, err := parseGroupLine(string(line))
		if err != nil {
			return nil, fmt.Errorf("group line %d: %w", lineno, err)
		}

		entries[entry.Name] = entry
	}

	return entries, nil //nolint:nilerr
}

func parseUint32(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(parsed), nil // ParseUint with the parameters specified guarantee the parsing of only 32 bits.
}
