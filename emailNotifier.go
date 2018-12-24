package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	// Fields to support E-Mail
	emailFromAddress    string
	emailToAddress      string
	emailServerHostname string
	emailServerPort     int
	emailAuthUsername   string
	emailAuthPassword   string
	emailAcceptAnyCerts bool
)

// NewEmailNotifier creates a new email notifier. Returns error if
// no valid email configuration is found
func NewEmailNotifier() (EmailNotifier, error) {
	// Only support new email configuration in v1.6+
	useNewConfigStyle()

	var notifier EmailNotifier

	// Note that some servers do not require authentication. We'll just
	// fail during send if authentication is required but not specified.
	if emailFromAddress == "" || emailToAddress == "" || emailServerHostname == "" || emailServerPort == 0 {
		return notifier, errors.New("Invalid E-mail configuration; Required fields missing")
	}

	return notifier, nil
}

// EmailNotifier handles email notifications as exposed by Notifier interface
type EmailNotifier struct{}

// NotifyOfStart is triggered when backup process starts
func (notifier EmailNotifier) NotifyOfStart() error {
	subject := fmt.Sprintf("duplicacy-util: Backup started for configuration %s", cmdConfig)
	return notifier.email(subject, []string{}, mailBody)
}

// NotifyOfSuccess is triggered when backup successfully finishes
func (notifier EmailNotifier) NotifyOfSuccess() error {
	subject := fmt.Sprintf("duplicacy-util: Backup results for configuration %s (success)", cmdConfig)
	return notifier.email(subject, htmlGenerateBody(), mailBody)
}

// NotifyOfSkip is triggered when a backup is skipped (if already running)
func (notifier EmailNotifier) NotifyOfSkip() error {
	subject := fmt.Sprintf("duplicacy-util: Backup results for configuration %s (skipped)", cmdConfig)
	return notifier.email(subject, htmlGenerateBody(), mailBody)
}

// NotifyOfFailure is triggered when a failure occurred during backup
func (notifier EmailNotifier) NotifyOfFailure() error {
	subject := fmt.Sprintf("duplicacy-util: Backup results for configuration %s (FAILURE)", cmdConfig)
	return notifier.email(subject, htmlGenerateBody(), mailBody)
}

// Email notification and return error if something went wrong
func (EmailNotifier) email(subject string, bodyHTML []string, bodyText []string) error {
	if err := sendMailMessage(subject, bodyHTML, bodyText); err != nil {
		// If an error occurred, we can't do much about it, so just log it (forcing output)
		quietFlag = false
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}
	return nil
}

func useNewConfigStyle() {
	emailFromAddress = viper.GetString("email.fromAddress")
	emailToAddress = viper.GetString("email.toAddress")
	emailServerHostname = viper.GetString("email.serverHostname")
	emailServerPort = viper.GetInt("email.serverPort")
	emailAuthUsername = viper.GetString("email.authUsername")

	// Allow environment variable DU_EMAIL_AUTH_PASSWORD to override authPassword in configuration
	if emailAuthPassword = os.Getenv("DU_EMAIL_AUTH_PASSWORD"); emailAuthPassword == "" {
		emailAuthPassword = viper.GetString("email.authPassword")
	}

	emailAcceptAnyCerts = viper.GetBool("email.acceptInsecureCerts")
}
