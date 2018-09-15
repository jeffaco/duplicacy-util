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
	"fmt"
	"html"
	"strings"

	"gopkg.in/gomail.v2"
)

const (
	htmlTableNone = 0 + iota
	htmlTableBackup
	htmlTableCopy
)

var (
	htmlTableContext int
)

func htmlGenerateBody() []string {
	// Construct the HTML mail body
	htmlBody := htmlConstructHeader()

	if len(backupTable) != 0 {
		htmlBody = append(htmlBody, htmlConstructTableBackupHeader()...)
		for _, entry := range backupTable {
			htmlBody = append(htmlBody, htmlContructTableBackupData(entry)...)
		}
		htmlBody = append(htmlBody, htmlConstructTableEnd()...)
	}

	if len(copyTable) != 0 {
		htmlBody = append(htmlBody, htmlConstructTableCopyHeader()...)
		for _, entry := range copyTable {
			htmlBody = append(htmlBody, htmlContructTableCopyData(entry)...)
		}
		htmlBody = append(htmlBody, htmlConstructTableEnd()...)
	}

	htmlBody = append(htmlBody, htmlConstructTrailer()...)

	return htmlBody
}

func htmlConstructHeader() []string {
	htmlTableContext = htmlTableNone

	return []string{
		`<!DOCTYPE html>`,
		`<html>`,
		`<head>`,
		`<style>`,
		`table {`,
		`    font-family: arial, sans-serif;`,
		`    border-collapse: collapse;`,
		`    width: 100%;`,
		`}`,
		`td, th {`,
		`    border: 1px solid #dddddd;`,
		`    text-align: right;`,
		`    padding: 8px;`,
		`}`,
		``,
		`tr:nth-child(even) {`,
		`    background-color: #dddddd;`,
		`}`,
		`</style>`,
		`</head>`,
		`<body>`,
		``,
		fmt.Sprintf(`<h1>Statistics for configuration: %s</h1>`, cmdConfig),
	}
}

func htmlConstructTableBackupHeader() []string {
	// Validate that our table context is correct
	if htmlTableContext != htmlTableNone {
		panic(fmt.Sprint("Invalid HTML Table Context: ", htmlTableContext))
	}

	htmlTableContext = htmlTableBackup

	return []string{
		``,
		`<h3>Backup Summary:</h3>`,
		`<table>`,
		`  <tr>`,
		`    <th style="text-align: left">Storage</th>`,
		`    <th>Duration</th>`,
		`    <th>Total Chunks</th>`,
		`	 <th>Total Used</th>`,
		`    <th>New Files</th>`,
		`    <th>New File Size</th>`,
		`	 <th>New Chunks</th>`,
		`	 <th>New Uploaded</th>`,
		`  </tr>`,
	}
}

func htmlContructTableBackupData(data backupRevision) []string {
	// Validate that our table context is correct
	if htmlTableContext != htmlTableBackup {
		panic(fmt.Sprint("Invalid HTML Table Context: ", htmlTableContext))
	}

	return []string{
		`  <tr>`,
		`    <td style="text-align: left">`, data.storage, `</td>`,
		`    <td>`, data.duration, `</td>`, // Like: "30:00:00"
		`    <td>`, data.chunkTotalCount, `</td>`, // Like: "348444"
		`    <td>`, data.chunkTotalSize, `</td>`, // Like: "1668G"
		`    <td>`, data.filesNewCount, `</td>`, // Like: "373"
		`    <td>`, data.filesNewSize, `</td>`, // Like: "15,951M"
		`    <td>`, data.chunkNewCount, `</td>`, // Like: "2415"
		`    <td>`, data.chunkNewUploaded, `</td>`, // Like: "12,255M"
		`  </tr>`,
	}
}

func htmlConstructTableCopyHeader() []string {
	// Validate that our table context is correct
	if htmlTableContext != htmlTableNone {
		panic(fmt.Sprint("Invalid HTML Table Context: ", htmlTableContext))
	}

	htmlTableContext = htmlTableCopy

	return []string{
		``,
		`<h3>Copy Summary:</h3>`,
		`<table>`,
		`  <tr>`,
		`    <th style="text-align: left">From Storage</th>`,
		`    <th style="text-align: left">To Storage</th>`,
		`    <th>Duration</th>`,
		`    <th>Total Chunks</th>`,
		`	 <th>Chunks Skipped</th>`,
		`	 <th>Chunks Copied</th>`,
		`  </tr>`,
	}
}

func htmlContructTableCopyData(data copyRevision) []string {
	// Validate that our table context is correct
	if htmlTableContext != htmlTableCopy {
		panic(fmt.Sprint("Invalid HTML Table Context: ", htmlTableContext))
	}

	return []string{
		`  <tr>`,
		`    <td style="text-align: left">`, data.storageFrom, `</td>`,
		`    <td style="text-align: left">`, data.storageTo, `</td>`,
		`    <td>`, data.duration, `</td>`,
		`    <td>`, data.chunkTotalCount, `</td>`,
		`    <td>`, data.chunkSkipCount, `</td>`,
		`    <td>`, data.chunkCopyCount, `</td>`,
		`  </tr>`,
	}
}

func htmlConstructTableEnd() []string {
	// Validate that our table context is correct
	if htmlTableContext == htmlTableNone {
		panic(fmt.Sprint("Invalid HTML Table Context: ", htmlTableContext))
	}

	htmlTableContext = htmlTableNone

	return []string{
		`</table>`,
	}
}

func htmlConstructTrailer() []string {
	var htmlMailBody []string

	for _, line := range mailBody {
		// Quoting is a little tricky here:
		// 1. We need to quote special HTML characters ('<', etc). Use 'html' pkg for that.
		// 2. We may have multiple spaces in strings, and we want to retain that.
		//    Since html.EscapeString doesn't do that, just do it ourselves.

		htmlMailBody = append(htmlMailBody, strings.Replace(html.EscapeString(line), " ", "&nbsp;", -1))
	}

	return []string{
		`</table>`,
		`<br><br><br><b>Log Text:</b><br><br>`,
		strings.Join(htmlMailBody, "<br>\n"),
		`</body>`,
		`</html>`,
	}
}

func sendMailMessage(subject string, bodyHTML []string, bodyText []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", emailFromAddress)
	m.SetHeader("To", emailToAddress)
	m.SetHeader("Subject", subject)
	m.SetBody("text", strings.Join(bodyText, "\r\n"))
	if len(bodyHTML) != 0 {
		m.AddAlternative("text/html", strings.Join(bodyHTML, "\r\n"))
	}

	d := gomail.NewDialer(emailServerHostname, emailServerPort, emailAuthUsername, emailAuthPassword)

	// Send the message
	return d.DialAndSend(m)
}
