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
	cmdFile   string
	cmdBackup bool
	cmdPurge  bool
	cmdCheck  bool

	configFile *utils.ConfigFile = utils.NewConfigFile()
)

func init() {
	// Perform command line argument processing
	flag.StringVar(&cmdFile, "file", "", "configuration file for storage definitions (must be specified)")
	flag.BoolVar(&cmdBackup, "backup", false, "perform duplicity backup operation")
	flag.BoolVar(&cmdPurge, "purge", false, "perform duplicity purge operation")
	flag.BoolVar(&cmdCheck, "check", false, "perform duplicity check operation")
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

	fmt.Println("Value for file:", cmdFile)
	fmt.Println("Value for backup:", cmdBackup)
	fmt.Println("Value for purge:", cmdPurge)
	fmt.Println("Value for check:", cmdCheck)

	// Parse the configuration file and check for errors
	configFile.SetConfig(cmdFile)
	if err := configFile.LoadConfig(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
