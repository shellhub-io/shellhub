package models

import (
	"gorm.io/gorm"
)

type Namespace struct {
	gorm.Model

	Name     string `json:"name"  validate:"required,hostname_rfc1123,excludes=.,lowercase"`
	TenantID string
}

type NamespaceSettings struct {
	SessionRecord          bool   `json:"session_record" bson:"session_record,omitempty"`
	ConnectionAnnouncement string `json:"connection_announcement" bson:"connection_announcement"`
}

type NamespaceChanges struct {
	Name                   string  `bson:"name,omitempty"`
	SessionRecord          *bool   `bson:"settings.session_record,omitempty"`
	ConnectionAnnouncement *string `bson:"settings.connection_announcement,omitempty"`
}

// default Announcement Message for the shellhub namespace
const DefaultAnnouncementMessage = `
******************************************************************
*                                                                *
*             Welcome to ShellHub Community Edition!             *
*                                                                *
* ShellHub is a next-generation SSH server, providing a          *
* seamless, secure, and user-friendly solution for remote        *
* access management. With ShellHub, you can manage all your      *
* devices effortlessly from a single platform, ensuring optimal  *
* security and productivity.                                     *
*                                                                *
* Want to learn more about ShellHub and explore other editions?  *
* Visit: https://shellhub.io                                     *
*                                                                *
* Join our community and contribute to our open-source project:  *
* https://github.com/shellhub-io/shellhub                        *
*                                                                *
* For assistance, please contact the system administrator.       *
*                                                                *
******************************************************************
`
