package envs

import (
	"fmt"
	"strings"
)

const announcementFormat = `
******************************************************************
*                                                                *
*%s*
*                                                                *
* ShellHub is a centralized SSH gateway for managing Linux       *
* devices and servers. Access any device from a web browser,     *
* mobile, or standard SSH client — no VPN, no public IP needed.  *
*                                                                *
%s* For assistance, contact the system administrator.              *
*                                                                *
******************************************************************
`

const communityLinks = `* Explore Cloud and Enterprise editions with advanced features:  *
* https://shellhub.io                                            *
*                                                                *
* Contribute to the open-source project:                         *
* https://github.com/shellhub-io/shellhub                       *
*                                                                *
`

// AnnouncementFor returns the SSH connection banner for the given edition.
func AnnouncementFor(edition Edition) string {
	if edition == "" {
		edition = Community
	}

	e := string(edition)
	title := "Welcome to ShellHub " + strings.ToUpper(e[:1]) + e[1:] + "!"
	pad := 64 - len(title)

	extra := ""
	if edition == Community {
		extra = communityLinks
	}

	return fmt.Sprintf(announcementFormat, fmt.Sprintf("%*s%s%*s", pad/2, "", title, pad-pad/2, ""), extra)
}
