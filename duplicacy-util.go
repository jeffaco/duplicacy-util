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
	"os"
	"path"

	"github.com/theckman/go-flock"
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
		fmt.Fprintln(os.Stderr, "Unrecognized arguments specified on command line:", flag.Args())
		os.Exit(2)
	}

	if cmdConfig == "" {
		fmt.Fprintln(os.Stderr, "Mandatory parameter -file is not specified (must be specified)")
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

	// Obtain a lock to make sure we don't overlap operations against a configuration
	lockfile := path.Join(globalLockDir, cmdConfig + ".lock")
	fileLock := flock.NewFlock(lockfile)

	locked, err := fileLock.TryLock()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(101)
	}

	if ! locked {
		// do not have exclusive lock
		err = errors.New("unable to obtain lock using lockfile: " + lockfile)
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(100)
	}

	// flock doesn't remove the lock file when done, so let's do it ourselves
	// (ignore any errors if we can't remove the lock file)
	defer os.Remove(lockfile)
	defer fileLock.Unlock()

	// Perform operations (backup or whatever)
	if err := performBackup(); err != nil {
		os.Exit(200)
	}

	os.Exit(0)
}

func performBackup() error {
	return nil
}
