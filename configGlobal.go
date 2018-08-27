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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// Location of duplicacy binary
	duplicacyPath string

	// Directory for lock files
	globalLockDir string

	// Directory for log files
	globalLogDir string

	// Number of log files to retain
	globalLogFileCount int

	// Fields to support E-Mail
	emailFromAddress    string
	emailToAddress      string
	emailServerHostname string
	emailServerPort     int
	emailAuthUsername   string
	emailAuthPassword   string
)

// loadGlobalConfig reads in config file and ENV variables if set.
func loadGlobalConfig(cfgFile string) error {
	var err error

	// Read in (or set) global environment variables
	if err = setGlobalConfigVariables(cfgFile); err != nil {
		return err
	}

	// Validate global environment variables
	if _, err = exec.LookPath(duplicacyPath); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}

	if err = verifyPathExists(globalLockDir); err != nil {
		return err
	}

	os.Mkdir(globalLogDir, 0755)
	if err = verifyPathExists(globalLogDir); err != nil {
		return err
	}

	if globalLogFileCount < 2 {
		err = errors.New("logfilecount must have at least two log files saved")
		fmt.Fprintln(os.Stderr, "Error:", err)
	}

	return nil
}

// Read configuration file or set reasonable defaults if none
func setGlobalConfigVariables(cfgFile string) error {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error", err)
		return err
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name "duplicacy-util" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath("$HOME/.duplicacy-util")
		viper.SetConfigName("duplicacy-util")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Set some defaults that we can depend on
	duplicacyPath = "duplicacy"
	globalLockDir = filepath.Join(home, ".duplicacy-util")
	globalLogDir = filepath.Join(home, ".duplicacy-util", "log")
	globalLogFileCount = 5

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// No configuration file is okay unless we specifically asked for a named file
		if cfgFile != "" {
			fmt.Fprintln(os.Stdout, "Error:", err)
			return err
		}
		return nil
	}

	logMessage(nil, fmt.Sprint("Using global config: ", viper.ConfigFileUsed()))

	if configStr := viper.GetString("duplicacypath"); configStr != "" {
		duplicacyPath = configStr
	}

	if configStr := viper.GetString("lockdirectory"); configStr != "" {
		globalLockDir = configStr
	}

	if configStr := viper.GetString("logdirectory"); configStr != "" {
		globalLogDir = configStr
	}

	if configInt := viper.GetInt("logfilecount"); configInt != 0 {
		globalLogFileCount = configInt
	}

	// No form of defaults for E-Mail settings, just read them in
	emailFromAddress = viper.GetString("emailFromAddress")
	emailToAddress = viper.GetString("emailToAddress")
	emailServerHostname = viper.GetString("emailServerHostname")
	emailServerPort = viper.GetInt("emailServerPort")
	emailAuthUsername = viper.GetString("emailAuthUsername")
	emailAuthPassword = viper.GetString("emailAuthPassword")

	return err
}

func verifyPathExists(path string) error {
	var err error

	if _, err = os.Stat(path); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		return err
	}

	return nil
}
