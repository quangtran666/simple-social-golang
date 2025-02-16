package mailer

import "embed"

const (
	FromName            = "Simple Social"
	MaxRetry            = 3
	UserWelcomeTemplate = "user_invitaion.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(username, email string) error
}
