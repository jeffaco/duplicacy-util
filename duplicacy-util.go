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
	"log"
	"os"
	"path/filepath"

	"github.com/theckman/go-flock"
	"time"
)

var (
	// Configuration file for backup operations
	cmdConfig string
	cmdGlobalConfig string

	// Binary options for what operations to perform
	cmdAll bool
	cmdBackup bool
	cmdCheck bool
	cmdPurge bool

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
	flag.BoolVar(&cmdPurge, "p", false, "Perform duplicacy purge operation")

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

	if cmdAll { cmdBackup, cmdPurge, cmdCheck = true, true, true }
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
	if !cmdBackup && !cmdPurge && !cmdCheck {
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
		return 101
	}

	if ! locked {
		// do not have exclusive lock
		err = errors.New("unable to obtain lock using lockfile: " + lockfile)
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 100
	}

	// flock doesn't remove the lock file when done, so let's do it ourselves
	// (ignore any errors if we can't remove the lock file)
	defer os.Remove(lockfile)
	defer fileLock.Unlock()

	// Perform operations (backup or whatever)
	if err := performBackup(); err != nil {
		return 200
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
				fmt.Println(time.Now().Format("15:04:05"), "Copying from storage", configFile.copyInfo[i]["from"],
					"to storage", configFile.copyInfo[i]["to"], "with", configFile.copyInfo[i]["threads"], "threads")
				err = Executor(duplicacyPath, cmdArgs, configFile.repoDir, anon)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error executing command:", err)
					logger.Println("Error executing command:", err)
					return err
				}
			}
		}
	}

	fmt.Println(time.Now().Format("15:04:05"), "Operations completed")

	return nil
}
