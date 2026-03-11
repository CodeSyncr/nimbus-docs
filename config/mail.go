/*
|--------------------------------------------------------------------------
| Mail Configuration
|--------------------------------------------------------------------------
|
| SMTP driver settings for outbound email.
|
*/

package config

var Mail MailConfig

type MailConfig struct {
	Driver string
	SMTP   SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

func loadMail() {
	Mail = MailConfig{
		Driver: env("MAIL_DRIVER", "smtp"),
		SMTP: SMTPConfig{
			Host:     env("SMTP_HOST", "localhost"),
			Port:     envInt("SMTP_PORT", 1025),
			Username: env("SMTP_USERNAME", ""),
			Password: env("SMTP_PASSWORD", ""),
			From:     env("MAIL_FROM", "noreply@example.com"),
			FromName: env("MAIL_FROM_NAME", "Nimbus App"),
		},
	}
}
