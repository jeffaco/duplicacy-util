package main

import (
	"errors"
	"fmt"

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
)

// NewEmailNotifier creates a new email notifier. Returns error if
// no valid email configuration is found
func NewEmailNotifier() (EmailNotifier, error) {
	usesNewConfigStyle := viper.IsSet("email")
	// check if old or new config style is being used
	if usesNewConfigStyle == true {
		useNewConfigStyle()
	} else {
		useOldConfigStyle()
	}

	var notifier EmailNotifier

	// Note that some servers do not require authentication. We'll just
	// fail during send if authentication is required but not specified.
	if emailFromAddress == "" || emailToAddress == "" || emailServerHostname == "" || emailServerPort == 0 {
		return notifier, errors.New("Invalid E-mail configuration; Required fields missing")
	}

	if usesNewConfigStyle == false {
		logMessage(nil, "Warning: E-Mail configuration in old format, update global configuration file")
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

// NotifyOfFailure is triggered when a failure occured during backup
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
	emailAuthPassword = viper.GetString("email.authPassword")
}

func useOldConfigStyle() {
	emailFromAddress = viper.GetString("emailFromAddress")
	emailToAddress = viper.GetString("emailToAddress")
	emailServerHostname = viper.GetString("emailServerHostname")
	emailServerPort = viper.GetInt("emailServerPort")
	emailAuthUsername = viper.GetString("emailAuthUsername")
	emailAuthPassword = viper.GetString("emailAuthPassword")
}
