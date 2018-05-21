package main

import (
	"strings"

	"gopkg.in/gomail.v2"
)

func sendTestMessage(subject string, body []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", emailFromAddress)
	m.SetHeader("To", emailToAddress)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", strings.Join(body, "\r\n"))

	d := gomail.NewDialer(emailServerHostname, emailServerPort, emailAuthUsername, emailAuthPassword)

	// Send the message
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
