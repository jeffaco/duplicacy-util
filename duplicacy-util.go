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
	"flag"
	"fmt"
	"os"
	"duplicacy-util/utils"
)

var (
	// Configuration file for backup operations
	cmdFile   string

	// Binary options for what operations to perform
	cmdAll bool
	cmdBackup bool
	cmdCheck  bool
	cmdPurge  bool

	debugFlag bool
	verboseFlag bool

	configFile *utils.ConfigFile = utils.NewConfigFile()
)

func init() {
	// Perform command line argument processing
	flag.StringVar(&cmdFile, "f", "", "Configuration file for storage definitions (must be specified)")

	flag.BoolVar(&cmdAll, "a", false, "Perform all duplicity operations (backup/copy, purge, check)")
	flag.BoolVar(&cmdBackup, "b", false, "Perform duplicity backup/copy operation")
	flag.BoolVar(&cmdCheck, "c", false, "Perform duplicity check operation")
	flag.BoolVar(&cmdPurge, "p", false, "Perform duplicity purge operation")

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

	if cmdFile == "" {
		fmt.Fprintln(os.Stderr, "Mandatory parameter -file is not specified (must be specified)")
		os.Exit(2)
	}

	if cmdAll { cmdBackup, cmdPurge, cmdCheck = true, true, true }
	if debugFlag { verboseFlag = true }

	// Parse the configuration file and check for errors
	configFile.SetConfig(cmdFile)
	if err := configFile.LoadConfig(verboseFlag, debugFlag); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
