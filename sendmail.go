// Copyright © 2018 Jeff Coffler <jeff@taltos.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"strings"

	"gopkg.in/gomail.v2"
)

func sendMailMessage(subject string, body []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", emailFromAddress)
	m.SetHeader("To", emailToAddress)
	m.SetHeader("Subject", subject)
	m.SetBody("text", strings.Join(body, "\r\n"))

	d := gomail.NewDialer(emailServerHostname, emailServerPort, emailAuthUsername, emailAuthPassword)

	// Send the message
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
