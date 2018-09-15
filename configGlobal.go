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

	// Notification publishers
	onSuccessNotifiers []Notifier
	onFailureNotifiers []Notifier
	onStartNotifiers   []Notifier
)

// loadGlobalConfig reads in config file and ENV variables if set.
func loadGlobalConfig(storageDir string, cfgFile string) error {
	var err error

	// Read in (or set) global environment variables
	if err = setGlobalConfigVariables(storageDir, cfgFile); err != nil {
		return err
	}

	// Validate global environment variables
	if _, err = exec.LookPath(duplicacyPath); err != nil {
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
func setGlobalConfigVariables(storageDir string, cfgFile string) error {
	// Reset config to prevent invalid state in case it's called multiple times
	viper.Reset()
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name "duplicacy-util" (without extension).
		viper.AddConfigPath(storageDir)
		viper.SetConfigName("duplicacy-util")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Set some defaults that we can depend on
	duplicacyPath = "duplicacy"
	globalLockDir = storageDir
	globalLogDir = filepath.Join(storageDir, "log")
	globalLogFileCount = 5
	onSuccessNotifiers = []Notifier{}
	onFailureNotifiers = []Notifier{}
	onStartNotifiers = []Notifier{}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// No configuration file is okay unless we specifically asked for a named file
		if cfgFile != "" {
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

	// Try to set default notifiers if non are defined in the config
	if viper.IsSet("notifications") == false {
		defaultNotifier, err := NewEmailNotifier()
		if err != nil {
			// No email configuration found
			return nil
		}
		onFailureNotifiers = append(onFailureNotifiers, defaultNotifier)
		onStartNotifiers = append(onStartNotifiers, defaultNotifier)
		return nil
	}

	var err error
	// Configure notifiers for onSucces notification
	if configSlice := viper.GetStringSlice("notifications.onSuccess"); len(configSlice) > 0 {
		onSuccessNotifiers, err = configureNotificationChannel(configSlice, "onSuccess")
		if err != nil {
			return err
		}
	}

	// Configure notifiers for onFailure notification
	if configSlice := viper.GetStringSlice("notifications.onFailure"); len(configSlice) > 0 {
		onFailureNotifiers, err = configureNotificationChannel(configSlice, "onFailure")
		if err != nil {
			return err
		}
	}

	// Configure notifiers for onStart notification
	if configSlice := viper.GetStringSlice("notifications.onStart"); len(configSlice) > 0 {
		onStartNotifiers, err = configureNotificationChannel(configSlice, "onStart")
		if err != nil {
			return err
		}
	}

	return nil
}

func configureNotificationChannel(channels []string, notificationType string) ([]Notifier, error) {
	notifiers := []Notifier{}
	for _, channel := range channels {
		if channel == "email" {
			emailNotifier, err := NewEmailNotifier()
			if err != nil {
				return nil, err
			}
			if isUniqueNotifier(emailNotifier, notifiers) {
				notifiers = append(notifiers, emailNotifier)
			}
			continue
		}
		// Return error if invalid notification channel is provided
		return nil, fmt.Errorf("Invalid notification channel \"%s\" provided for %sNotifier", channel, notificationType)
	}
	return notifiers, nil
}

func verifyPathExists(path string) error {
	var err error

	if _, err = os.Stat(path); err != nil {
		return err
	}

	return nil
}

func hasFailureNotifier() bool {
	return len(onFailureNotifiers) > 0
}

// Checks (by type) if notifier is unique
func isUniqueNotifier(notifier Notifier, collection []Notifier) bool {
	for _, existing := range collection {
		if notifier.(Notifier) == existing.(Notifier) {
			return false
		}
	}
	return true
}
