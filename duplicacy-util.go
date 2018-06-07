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
	"regexp"
	"strings"
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
	mailBody    []string

	// Create configuration object to load configuration file
	configFile *ConfigFile = NewConfigFile()
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
	mailBody = append(mailBody, text)

	if !quietFlag {
		if w == os.Stdout {
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
		backupRecord := backupRevision{
			storage:          "test",
			chunkTotalCount:  "384444",
			chunkTotalSize:   "1668G",
			filesTotalCount:  "161318",
			filesTotalSize:   "1666G",
			filesNewCount:    "373",
			filesNewSize:     "15,951M",
			chunkNewCount:    "2415",
			chunkNewSize:     "12,391M",
			chunkNewUploaded: "12,255M",
			duration:         "30:05",
		}
		backupTable = append(backupTable, backupRecord)
		cmdConfig = "test"

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

func performBackup() error {
	// Handle log file rotation (before any output to log file so old one doesn't get trashed)

	logMessage(nil, "Rotating log files")
	if err := rotateLogFiles(); err != nil {
		return err
	}

	// Create output log file
	file, err := os.Create(filepath.Join(globalLogDir, cmdConfig+".log"))
	if err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}
	logger := log.New(file, "", log.Ltime)

	startTime := time.Now()

	logMessage(logger, fmt.Sprint("Beginning backup on ", time.Now().Format("01-02-2006 15:04:05")))

	// Handling when processing output from "duplicacy backup" command
	var backupEntry backupRevision

	backupLogger := func(line string) {
		switch {
		// Files: 161318 total, 1666G bytes; 373 new, 15,951M bytes
		case strings.HasPrefix(line, "Files:"):
			logger.Println(line)
			logMessage(logger, fmt.Sprint("  ", line))

			// Save chunk data for inclusion into HTML portion of E-Mail message
			re := regexp.MustCompile(`.*: (\S+) total, (\S+) bytes; (\S+) new, (\S+) bytes`)
			elements := re.FindStringSubmatch(line)
			if len(elements) >= 4 {
				backupEntry.filesTotalCount = elements[1]
				backupEntry.filesTotalSize = elements[2]
				backupEntry.filesNewCount = elements[3]
				backupEntry.filesNewSize = elements[4]
			}

			// All chunks: 348444 total, 1668G bytes; 2415 new, 12,391M bytes, 12,255M bytes uploaded
		case strings.HasPrefix(line, "All chunks:"):
			logger.Println(line)
			logMessage(logger, fmt.Sprint("  ", line))

			// Save chunk data for inclusion into HTML portion of E-Mail message
			re := regexp.MustCompile(`.*: (\S+) total, (\S+) bytes; (\S+) new, (\S+) bytes, (\S+) bytes uploaded`)
			elements := re.FindStringSubmatch(line)
			if len(elements) >= 6 {
				backupEntry.chunkTotalCount = elements[1]
				backupEntry.chunkTotalSize = elements[2]
				backupEntry.chunkNewCount = elements[3]
				backupEntry.chunkNewSize = elements[4]
				backupEntry.chunkNewUploaded = elements[5]
			}
		default:
			logger.Println(line)
		}
	}

	copyLogger := func(line string) {
		switch {
		// Copy complete, 107 total chunks, 0 chunks copied, 107 skipped
		case strings.HasPrefix(line, "Copy complete, "):
			logger.Println(line)
			logMessage(logger, fmt.Sprint("  ", line))
		default:
			logger.Println(line)
		}
	}

	// Handling when processing output from generic "duplicacy" command
	anon := func(s string) { logger.Println(s) }

	// Perform backup/copy operations if requested
	if cmdBackup {
		for i := range configFile.backupInfo {
			backupStartTime := time.Now()
			logger.Println("######################################################################")
			cmdArgs := []string{"backup", "-storage", configFile.backupInfo[i]["name"], "-threads", configFile.backupInfo[i]["threads"], "-stats"}
			logMessage(logger, fmt.Sprint("Backing up to storage ", configFile.backupInfo[i]["name"],
				" with ", configFile.backupInfo[i]["threads"], " threads"))
			if debugFlag {
				logMessage(logger, fmt.Sprint("Executing: ", duplicacyPath, cmdArgs))
			}
			err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, backupLogger)
			if err != nil {
				logError(logger, fmt.Sprint("Error executing command: ", err))
				return err
			}
			backupDuration := getTimeDiffString(backupStartTime, time.Now())
			logMessage(logger, fmt.Sprint("  Duration: ", backupDuration))

			// Save data from backup for HTML table in E-Mail
			backupEntry.storage = configFile.backupInfo[i]["name"]
			backupEntry.duration = backupDuration
			backupTable = append(backupTable, backupEntry)
		}
		if len(configFile.copyInfo) != 0 {
			copyStartTime := time.Now()
			for i := range configFile.copyInfo {
				logger.Println("######################################################################")
				cmdArgs := []string{"copy", "-threads", configFile.copyInfo[i]["threads"],
					"-from", configFile.copyInfo[i]["from"], "-to", configFile.copyInfo[i]["to"]}
				logMessage(logger, fmt.Sprint("Copying from storage ", configFile.copyInfo[i]["from"],
					" to storage ", configFile.copyInfo[i]["to"], " with ", configFile.copyInfo[i]["threads"], " threads"))
				if debugFlag {
					logMessage(logger, fmt.Sprint("Executing: ", duplicacyPath, cmdArgs))
				}
				err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, copyLogger)
				if err != nil {
					logError(logger, fmt.Sprint("Error executing command: ", err))
					return err
				}
				logMessage(logger, fmt.Sprint("  Duration: ", getTimeDiffString(copyStartTime, time.Now())))
			}
		}
	}

	// Perform prune operations if requested
	if cmdPrune {
		for i := range configFile.pruneInfo {
			logger.Println("######################################################################")
			cmdArgs := []string{"prune", "-all", "-storage", configFile.pruneInfo[i]["storage"]}
			cmdArgs = append(cmdArgs, strings.Split(configFile.pruneInfo[i]["keep"], " ")...)
			logMessage(logger, fmt.Sprint("Pruning storage ", configFile.pruneInfo[i]["storage"]))
			if debugFlag {
				logMessage(logger, fmt.Sprint("Executing: ", duplicacyPath, cmdArgs))
			}
			err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, anon)
			if err != nil {
				logError(logger, fmt.Sprint("Error executing command: ", err))
				return err
			}
		}
	}

	// Perform check operations if requested
	if cmdCheck {
		for i := range configFile.checkInfo {
			logger.Println("######################################################################")
			cmdArgs := []string{"check", "-storage", configFile.checkInfo[i]["storage"]}
			if configFile.checkInfo[i]["all"] == "true" {
				cmdArgs = append(cmdArgs, "-all")
			}
			logMessage(logger, fmt.Sprint("Checking storage ", configFile.pruneInfo[i]["storage"]))
			if debugFlag {
				logMessage(logger, fmt.Sprint("Executing: ", duplicacyPath, cmdArgs))
			}
			err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, anon)
			if err != nil {
				logError(logger, fmt.Sprint("Error executing command: ", err))
				return err
			}
		}
	}

	endTime := time.Now()

	logger.Println("######################################################################")
	logMessage(logger, fmt.Sprint("Operations completed in ", getTimeDiffString(startTime, endTime)))

	return nil
}
