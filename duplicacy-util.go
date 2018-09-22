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

var (
	// Configuration file for backup operations
	cmdConfig       string // Name of the configuration file for repository
	cmdGlobalConfig string // Name of the global configuration file (normally "duplicacy-util")
	cmdStorageDir   string // Base directory for storage of global/repository/log files

	// Binary options for what operations to perform
	cmdAll    bool
	cmdBackup bool
	cmdCheck  bool
	cmdPrune  bool

	sendMail              bool
	testMailFlag          bool
	testNotificationsFlag bool

	debugFlag   bool
	quietFlag   bool
	verboseFlag bool
	versionFlag bool

	// Version flags (passed by link stage)
	versionText = "<dev>"
	gitHash     = "<unknown>"

	// Mail message body to send upon completion
	backupTable []backupRevision
	copyTable   []copyRevision
	mailBody    []string

	// Create configuration object to load configuration file
	configFile = newConfigurationFile()

	// Display time in local output messages?
	loggingSystemDisplayTime = true

	// Global storage directory (location where all files are stored)
	globalStorageDirectory string

	// Unit testing active?
	runningUnitTests bool
)

func init() {
	// Perform command line argument processing
	flag.StringVar(&cmdConfig, "f", "", "Configuration file for storage definitions (must be specified)")
	flag.StringVar(&cmdGlobalConfig, "g", "", "Global configuration file name")
	flag.StringVar(&cmdStorageDir, "sd", "", "Full path to storage directory for configuration/log files")

	flag.BoolVar(&cmdAll, "a", false, "Perform all duplicacy operations (backup/copy, purge, check)")
	flag.BoolVar(&cmdBackup, "b", false, "Perform duplicacy backup/copy operation")
	flag.BoolVar(&cmdCheck, "c", false, "Perform duplicacy check operation")
	flag.BoolVar(&cmdPrune, "p", false, "Perform duplicacy prune operation")

	flag.BoolVar(&sendMail, "m", false, "(Deprecated) Send E-Mail with results of operations (implies quiet)")
	flag.BoolVar(&testMailFlag, "tm", false, "(Deprecated: Use -tn instead) Send a test message via E-Mail")
	flag.BoolVar(&testNotificationsFlag, "tn", false, "Test notifications")

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
	var err error

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

	// If version number was requested, show it and exit
	if versionFlag {
		fmt.Printf("Version: %s, Git Hash: %s\n", versionText, gitHash)
		os.Exit(0)
	}

	// Determine the location of the global storage directory
	globalStorageDirectory, err = getStorageDirectory(cmdStorageDir)
	if err != nil {
		os.Exit(2)
	}

	// Parse the global configuration file, if any
	if err := loadGlobalConfig(globalStorageDirectory, cmdGlobalConfig); err != nil {
		quietFlag = false
		logError(nil, fmt.Sprintf("Error: %s", err))
		os.Exit(2)
	}

	returnStatus, err := processArguments()
	if err != nil {
		switch returnStatus {
		case 6200:
			// Notify that the backup process has been skipped
			logError(nil, fmt.Sprintf("Warning: %s", err))
			notifyOfSkip()

		default:
			// Notify that the backup process has failed
			logError(nil, fmt.Sprintf("Error: %s", err))
			notifyOfFailure()
		}
	}

	os.Exit(returnStatus)
}

func processArguments() (int, error) {

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

	// if no failure notifier is defined quiet mode is not allowed
	if quietFlag && hasFailureNotifier() == false {
		quietFlag = false
		logError(nil, "Notice: Quiet mode refused; a failure notifier should be configured")
	}

	// Handle request to test Notifications
	// if testmailFlag is set; only email notifications will be tested
	if testNotificationsFlag || testMailFlag {
		return 1, testNotifications()
	}

	if cmdConfig == "" {
		return 2, errors.New("Mandatory parameter -f is not specified (must be specified)")
	}

	// Parse the configuration file and check for errors
	// (Errors are printed to stderr as well as returned)
	configFile.setConfig(cmdConfig)
	if err := configFile.loadConfig(verboseFlag, debugFlag); err != nil {
		return 1, nil
	}

	// Everything is loaded; make sure we hae something to do
	if !cmdBackup && !cmdPrune && !cmdCheck {
		return 1, errors.New("No operations to perform (specify -b, -p, -c, or -a)")
	}

	// Perform processing. Note that int is returned for two reasons:
	// 1. We need to know the proper exit code
	// 2. We want defer statements to execute, so we can't use os.Exit here

	logMessage(nil, fmt.Sprintf("duplicacy-util starting, version: %s, Git Hash: %s", versionText, gitHash))
	return obtainLock()
}

func obtainLock() (int, error) {
	// Obtain a lock to make sure we don't overlap operations against a configuration
	lockfile := filepath.Join(globalLockDir, cmdConfig+".lock")
	fileLock := flock.NewFlock(lockfile)

	locked, err := fileLock.TryLock()
	if err != nil {
		return 201, err
	}

	if !locked {
		// do not have exclusive lock
		return 6200, errors.New("Backup already running and will be skipped")
	}

	// flock doesn't remove the lock file when done, so let's do it ourselves
	// (ignore any errors if we can't remove the lock file)
	defer os.Remove(lockfile)
	defer fileLock.Unlock()

	// Perform operations (backup or whatever)
	if err := performBackup(); err != nil {
		return 500, errors.New("Backup failed. Check the logs for details")
	}

	return 0, nil
}
