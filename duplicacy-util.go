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
	"compress/gzip"
	"errors"
	"io"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/theckman/go-flock"
)

var (
	// Configuration file for backup operations
	cmdConfig string
	cmdGlobalConfig string

	// Binary options for what operations to perform
	cmdAll    bool
	cmdBackup bool
	cmdCheck  bool
	cmdPrune  bool

	debugFlag bool
	verboseFlag bool

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

	flag.BoolVar(&debugFlag, "d", false, "Enable debug output (implies verbose)")
	flag.BoolVar(&verboseFlag, "v", false, "Enable verbose output")
}

func main() {
	// Parse the command line arguments and validate results
	flag.Parse()

	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Error: Unrecognized arguments specified on command line:", flag.Args())
		os.Exit(2)
	}

	if cmdConfig == "" {
		fmt.Fprintln(os.Stderr, "Error: Mandatory parameter -file is not specified (must be specified)")
		os.Exit(2)
	}

	if cmdAll { cmdBackup, cmdPrune, cmdCheck = true, true, true }
	if debugFlag { verboseFlag = true }

	// Parse the global configuration file, if any
	if err := loadGlobalConfig(cmdGlobalConfig); err != nil {
		os.Exit(2)
	}

	// Parse the configuration file and check for errors
	// (Errors are printed to stderr as well as returned)
	configFile.SetConfig(cmdConfig)
	if err := configFile.LoadConfig(verboseFlag, debugFlag); err != nil {
		os.Exit(1)
	}

	// Everything is loaded; make sure we hae something to do
	if !cmdBackup && !cmdPrune && !cmdCheck {
		fmt.Fprintln(os.Stderr, "Error: No operations to perform (specify -b, -p, -c, or -a)")
		os.Exit(1)
	}

	// Perform processing. Note that int is returned for two reasons:
	// 1. We need to know the proper exit code
	// 2. We want defer statements to execute, so we only use os.Exit here

	os.Exit( obtainLock() )
}

func obtainLock() int {
	// Obtain a lock to make sure we don't overlap operations against a configuration
	lockfile := filepath.Join(globalLockDir, cmdConfig + ".lock")
	fileLock := flock.NewFlock(lockfile)

	locked, err := fileLock.TryLock()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 201
	}

	if ! locked {
		// do not have exclusive lock
		err = errors.New("unable to obtain lock using lockfile: " + lockfile)
		fmt.Fprintln(os.Stderr, "Error:", err)
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
	// Create output log file
	file, err := os.Create(filepath.Join(globalLogDir, cmdConfig + ".log"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}
	logger := log.New(file, "", log.Ltime)

	logger.Println("Beginning backup on", time.Now().Format("01-02-2006 15:04:05"))
	fmt.Println(time.Now().Format("15:04:05"), "Beginning backup on", time.Now().Format("01-02-2006 15:04:05"))

	// Handle log file rotation

	logger.Println("Rotating log files")
	fmt.Println(time.Now().Format("15:04:05"), "Rotating log files")

	if err := rotateLogFiles(); err != nil {
		return err
	}


	anon := func(s string) { logger.Println(s) }

	// Perform backup/copy operations if requested
	if cmdBackup {
		for i := range configFile.backupInfo {
			logger.Println("######################################################################")
			cmdArgs := []string{"backup", "-storage", configFile.backupInfo[i]["name"], "-threads", configFile.backupInfo[i]["threads"], "-stats"}
			logger.Println("Backing up to storage", configFile.backupInfo[i]["name"],
				"with", configFile.backupInfo[i]["threads"], "threads")
			fmt.Println(time.Now().Format("15:04:05"), "Backing up to storage", configFile.backupInfo[i]["name"],
				"with", configFile.backupInfo[i]["threads"], "threads")
			if debugFlag { fmt.Println("Executing:", duplicacyPath, cmdArgs) }
			err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, anon)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error executing command:", err)
				logger.Println("Error executing command:", err)
				return err
			}
		}
		if len(configFile.copyInfo) != 0 {
			for i := range configFile.copyInfo {
				logger.Println("######################################################################")
				cmdArgs := []string{"copy", "-threads", configFile.copyInfo[i]["threads"],
					"-from", configFile.copyInfo[i]["from"], "-to", configFile.copyInfo[i]["to"]}
				logger.Println("Copying from storage", configFile.copyInfo[i]["from"],
					"to storage", configFile.copyInfo[i]["to"], "with", configFile.copyInfo[i]["threads"], "threads")
				fmt.Println(time.Now().Format("15:04:05"), "Copying from storage", configFile.copyInfo[i]["from"],
					"to storage", configFile.copyInfo[i]["to"], "with", configFile.copyInfo[i]["threads"], "threads")
				if debugFlag { fmt.Println("Executing:", duplicacyPath, cmdArgs) }
				err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, anon)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error executing command:", err)
					logger.Println("Error executing command:", err)
					return err
				}
			}
		}
	}

	// Perform prune operations if requested
	if cmdPrune {
		for i := range configFile.pruneInfo {
			logger.Println("######################################################################")
			cmdArgs := []string{"prune", "-all", "-storage", configFile.pruneInfo[i]["storage"]}
			cmdArgs = append(cmdArgs, strings.Split(configFile.pruneInfo[i]["keep"], " ")...)
			logger.Println("Pruning storage", configFile.pruneInfo[i]["storage"])
			fmt.Println(time.Now().Format("15:04:05"), "Pruning storage", configFile.pruneInfo[i]["storage"])
			if debugFlag { fmt.Println("Executing:", duplicacyPath, cmdArgs) }
			err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, anon)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error executing command:", err)
				logger.Println("Error executing command:", err)
				return err
			}
		}
	}

	// Perform check operations if requested
	if cmdCheck {
		for i := range configFile.checkInfo {
			logger.Println("######################################################################")
			cmdArgs := []string{"check", "-storage", configFile.checkInfo[i]["storage"]}
			if configFile.checkInfo[i]["all"] == "true" { cmdArgs = append(cmdArgs, "-all") }
			logger.Println("Checking storage", configFile.pruneInfo[i]["storage"])
			fmt.Println(time.Now().Format("15:04:05"), "Checking storage", configFile.pruneInfo[i]["storage"])
			if debugFlag { fmt.Println("Executing:", duplicacyPath, cmdArgs) }
			err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, anon)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error executing command:", err)
				logger.Println("Error executing command:", err)
				return err
			}
		}
	}

	fmt.Println(time.Now().Format("15:04:05"), "Operations completed")

	return nil
}

func rotateLogFiles() error {
	logFileRoot := filepath.Join(globalLogDir, cmdConfig) + ".log"

	// Kick the older log files up by a count of one
	for i := globalLogFileCount - 2; i >= 1; i-- {
		os.Rename(logFileRoot + "." + strconv.Itoa(i) + ".gz",
			logFileRoot + "." + strconv.Itoa(i+1) + ".gz")
	}

	// If uncompressed log file exists, rename it and compress it
	if _, err := os.Stat(logFileRoot); os.IsNotExist(err) {
		return nil
	}

	// Compress <file.log> into <file.log.1.gz>
	reader, err := os.Open(logFileRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}

	writer, err := os.Create(logFileRoot + ".1.gz")
	if err != nil {
		reader.Close()
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = logFileRoot + ".1.gz"
	defer archiver.Close()

	if _, err := io.Copy(archiver, reader); err != nil {
		panic(err)
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}

	return nil
}
