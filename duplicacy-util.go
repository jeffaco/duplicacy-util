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
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/theckman/go-flock"
)

type backupRevision struct {
	storage          string
	chunkTotalCount  string // Like: 348444
	chunkTotalSize   string // Like: 1668G
	filesTotalCount  string // Like: 161318
	filesTotalSize   string // Like: 1666G
	filesNewCount    string // Like: 373
	filesNewSize     string // Like: 15,951M
	chunkNewCount    string // Like: 2415
	chunkNewSize     string // Like: 12,391M
	chunkNewUploaded string // Like: 12,255M
	duration         string
}

type copyRevision struct {
	storageFrom     string
	storageTo       string // Like: 348444
	chunkTotalCount string // Like: 109
	chunkCopyCount  string // Like: 3
	chunkSkipCount  string // Like: 106
	duration        string
}

var (
	// Configuration file for backup operations
	cmdConfig       string
	cmdGlobalConfig string

	// Binary options for what operations to perform
	cmdAll    bool
	cmdBackup bool
	cmdCheck  bool
	cmdPrune  bool

	sendMail bool
	testMail bool

	debugFlag   bool
	quietFlag   bool
	verboseFlag bool
	versionFlag bool

	// Version flags (passed by link stage)
	versionText string = "<dev>"
	gitHash     string = "<unknown>"

	// Mail message body to send upon completion
	backupTable []backupRevision
	copyTable   []copyRevision
	mailBody    []string

	// Create configuration object to load configuration file
	configFile *ConfigFile = NewConfigFile()

	// Display time in local output messages?
	loggingSystemDisplayTime bool = true
)

func init() {
	// Perform command line argument processing
	flag.StringVar(&cmdConfig, "f", "", "Configuration file for storage definitions (must be specified)")

	flag.BoolVar(&cmdAll, "a", false, "Perform all duplicacy operations (backup/copy, purge, check)")
	flag.BoolVar(&cmdBackup, "b", false, "Perform duplicacy backup/copy operation")
	flag.BoolVar(&cmdCheck, "c", false, "Perform duplicacy check operation")
	flag.StringVar(&cmdGlobalConfig, "g", "", "Global configuration file name")
	flag.BoolVar(&cmdPrune, "p", false, "Perform duplicacy prune operation")

	flag.BoolVar(&sendMail, "m", false, "Send E-Mail with results of operations (implies quiet)")
	flag.BoolVar(&testMail, "tm", false, "Send a test message via E-Mail")

	flag.BoolVar(&debugFlag, "d", false, "Enable debug output (implies verbose)")
	flag.BoolVar(&quietFlag, "q", false, "Quiet operations (generate output only in case of error)")
	flag.BoolVar(&verboseFlag, "v", false, "Enable verbose output")
	flag.BoolVar(&versionFlag, "version", false, "Display version number")
}

// Generic output routine to generate output to screen (and E-Mail) - Allow output writer
func logFMessage(w io.Writer, logger *log.Logger, message string) {
	if logger != nil {
		logger.Println(message)
	}

	text := fmt.Sprint(time.Now().Format("15:04:05"), " ", message)
	if loggingSystemDisplayTime == false {
		text = message
	}
	mailBody = append(mailBody, text)

	if !quietFlag {
		if w == os.Stdout && loggingSystemDisplayTime == true {
			fmt.Fprintln(w, text)
		} else {
			// Fatal message shouldn't have time prefix
			fmt.Fprintln(w, message)
		}
	}
}

// Generic error output routine to generate output to screen (and E-Mail)
func logError(logger *log.Logger, message string) {
	logFMessage(os.Stderr, logger, message)
}

// Generic output routine to generate output to screen (and E-Mail)
func logMessage(logger *log.Logger, message string) {
	logFMessage(os.Stdout, logger, message)
}

func main() {
	// Parse the command line arguments and validate results
	flag.Parse()

	// We do minimal command line processing here. Just things we KNOW
	// won't be supported via automated launching. Otherwise, send off
	// to processor so we can capture as much as possible via E-Mail
	// if so configured.

	if flag.NArg() != 0 {
		logError(nil, fmt.Sprint("Error: Unrecognized arguments specified on command line: ", flag.Args()))
		os.Exit(2)
	}

	// Parse the global configuration file, if any
	if err := loadGlobalConfig(cmdGlobalConfig); err != nil {
		os.Exit(2)
	}

	// If version number was requested, show it and exit
	if versionFlag {
		fmt.Printf("Version: %s, Git Hash: %s\n", versionText, gitHash)
		os.Exit(0)
	}

	returnStatus, transmitMail := processArguments()

	// Send mail if we were requested to do so
	if transmitMail {
		var ind string = "(success)"
		if returnStatus != 0 {
			ind = "(FAILURE)"
		}
		subject := fmt.Sprintf("duplicacy-util: Backup results for configuration %s %s", cmdConfig, ind)

		// Send the mail message
		if err := sendMailMessage(subject, htmlGenerateBody(), mailBody); err != nil {
			// If an error occurred, we can't do much about it, so just log it (forcing output)
			quietFlag = false
			logError(nil, fmt.Sprint("Error: ", err))
		}
	}

	os.Exit(returnStatus)
}

func processArguments() (int, bool) {
	var transmitMail bool = false

	if cmdAll {
		cmdBackup, cmdPrune, cmdCheck = true, true, true
	}
	if debugFlag {
		verboseFlag = true
	}

	// Verbose overrides quiet
	if verboseFlag == true && quietFlag == true {
		quietFlag = false
	}

	// Handle request to send test E-Mail, if requested
	if testMail {
		cmdConfig = "test"

		backupTable = []backupRevision{
			backupRevision{
				storage:          "b2",
				chunkTotalCount:  "149",
				chunkTotalSize:   "870,624K",
				filesTotalCount:  "345",
				filesTotalSize:   "823,261K",
				filesNewCount:    "1",
				filesNewSize:     "7,984K",
				chunkNewCount:    "6",
				chunkNewSize:     "8,106K",
				chunkNewUploaded: "3,410K",
				duration:         "9 seconds",
			},
			backupRevision{
				storage:          "azure-direct",
				chunkTotalCount:  "149",
				chunkTotalSize:   "870,624K",
				filesTotalCount:  "345",
				filesTotalSize:   "823,261K",
				filesNewCount:    "1",
				filesNewSize:     "7,984K",
				chunkNewCount:    "6",
				chunkNewSize:     "8,106K",
				chunkNewUploaded: "3,410K",
				duration:         "2 seconds",
			},
		}

		copyTable = []copyRevision{
			copyRevision{
				storageFrom:     "b2",
				storageTo:       "azure-direct",
				chunkTotalCount: "109",
				chunkCopyCount:  "3",
				chunkSkipCount:  "106",
				duration:        "9 seconds",
			},
		}

		if err := sendMailMessage("duplicacy-util: Backup results for configuration test (success)",
			htmlGenerateBody(),
			[]string{"This is a test E-Mail message for a successful backup job"}); err != nil {
			fmt.Fprintln(os.Stderr, "Error sending succcess E-Mail message:", err)
		}

		if err := sendMailMessage("duplicacy-util: Backup results for configuration test (FAILURE)",
			htmlGenerateBody(),
			[]string{"This is a test E-Mail message for a failed backup job"}); err != nil {
			fmt.Fprintln(os.Stderr, "Error sending failed E-Mail message:", err)
		}

		return 1, transmitMail
	}

	// Basic handling for E-Mail; only honor it if it's configured
	// (If it's not, disallow quiet operations or we won't see errors)
	if sendMail {
		if emailFromAddress == "" || emailToAddress == "" || emailServerHostname == "" || emailServerPort == 0 ||
			emailAuthUsername == "" || emailAuthPassword == "" {
			quietFlag = false
			logError(nil, "Error: Unable to send E-Mail; required fields missing from global configuration")
			return 3, transmitMail
		}

		transmitMail = true
	} else {
		if quietFlag {
			quietFlag = false
			logError(nil, "Notice: Quiet mode refused; makes no sense without sending mail")
		}
	}

	if cmdConfig == "" {
		logError(nil, "Error: Mandatory parameter -f is not specified (must be specified)")
		return 2, transmitMail
	}

	// Parse the configuration file and check for errors
	// (Errors are printed to stderr as well as returned)
	configFile.SetConfig(cmdConfig)
	if err := configFile.LoadConfig(verboseFlag, debugFlag); err != nil {
		return 1, transmitMail
	}

	// Everything is loaded; make sure we hae something to do
	if !cmdBackup && !cmdPrune && !cmdCheck {
		logError(nil, "Error: No operations to perform (specify -b, -p, -c, or -a)")
		return 1, transmitMail
	}

	// Perform processing. Note that int is returned for two reasons:
	// 1. We need to know the proper exit code
	// 2. We want defer statements to execute, so we can't use os.Exit here

	logMessage(nil, fmt.Sprintf("duplicacy-util starting, version: %s, Git Hash: %s", versionText, gitHash))
	return obtainLock(), transmitMail
}

func obtainLock() int {
	// Obtain a lock to make sure we don't overlap operations against a configuration
	lockfile := filepath.Join(globalLockDir, cmdConfig+".lock")
	fileLock := flock.NewFlock(lockfile)

	locked, err := fileLock.TryLock()
	if err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return 201
	}

	if !locked {
		// do not have exclusive lock
		err = errors.New("unable to obtain lock using lockfile: " + lockfile)
		logError(nil, fmt.Sprint("Error: ", err))
		return 200
	}

	// flock doesn't remove the lock file when done, so let's do it ourselves
	// (ignore any errors if we can't remove the lock file)
	defer os.Remove(lockfile)
	defer fileLock.Unlock()

	// Perform operations (backup or whatever)
	if err := performBackup(); err != nil {
		return 500
	}

	return 0
}
