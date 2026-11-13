package command

import (
	"strconv"
	"strings"
)

// FormatSFTPGroups encodes the supplementary groups an SFTP server process must run with, in the
// form [ParseSFTPGroups] reads. An empty or nil list encodes as the empty string.
func FormatSFTPGroups(groups []uint32) string {
	encoded := make([]string, 0, len(groups))
	for _, group := range groups {
		encoded = append(encoded, strconv.FormatUint(uint64(group), 10))
	}

	return strings.Join(encoded, ",")
}

// ParseSFTPGroups decodes a list written by [FormatSFTPGroups]. The empty string decodes to an
// empty, non-nil list. It returns an error when an entry is not a decimal group ID that fits in 32
// bits, so a malformed list is refused instead of being read as fewer groups.
func ParseSFTPGroups(value string) ([]int, error) {
	groups := []int{}
	if value == "" {
		return groups, nil
	}

	for entry := range strings.SplitSeq(value, ",") {
		group, err := strconv.ParseUint(entry, 10, 32)
		if err != nil {
			return nil, err
		}

		groups = append(groups, int(group))
	}

	return groups, nil
}
