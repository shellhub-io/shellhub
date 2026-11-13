package host

import (
	"os"
	"path"
	"strings"
)

var confiningShells = map[string]struct{}{
	"nologin":   {},
	"false":     {},
	"git-shell": {},
	"rbash":     {},
	"rksh":      {},
	"rzsh":      {},
}

func shellConfinesSFTP(shell string) bool {
	if strings.EqualFold(os.Getenv("SHELLHUB_SFTP_ALLOW_RESTRICTED_SHELLS"), "true") {
		return false
	}

	_, confining := confiningShells[path.Base(shell)]

	return confining
}
