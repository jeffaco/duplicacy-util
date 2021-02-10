// Copyright © 2018 Jeff Coffler <jeff@taltos.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var oldBackupFileFormat = false

type configurationFile struct {
	// Name (without extension) of the configuration file
	configFilename string

	// Directory for repository
	repoDir string

	// Storage information for backup, copy, prune, and check commands respectively
	backupInfo []map[string]string
	copyInfo   []map[string]string
	pruneInfo  []map[string]string
	checkInfo  []map[string]string
}

func newConfigurationFile() *configurationFile {
	config := new(configurationFile)
	return config
}

func (config *configurationFile) setConfig(cnfFile string) {
	config.configFilename = cnfFile
}

func (config *configurationFile) loadConfig(verboseFlag bool, debugFlag bool) error {
	var err error

	// Separate config file should use new viper instance
	v := viper.New()

	// Search config in home directory with name ".duplicacy-util" (without extension).
	v.AddConfigPath(globalStorageDirectory)
	v.SetConfigName(config.configFilename)

	v.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}

	// Grab the repository location
	config.repoDir = v.GetString("repository")
	if config.repoDir == "" {
		err = errors.New("missing mandatory repository location")
		logError(nil, fmt.Sprint("Error: ", err))
	}
	if _, err = os.Stat(config.repoDir); err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
	}

	// Populate information from configuration
	config.backupInfo = readSection(v, config.configFilename, "storage")
	config.copyInfo = readSection(v, config.configFilename, "copy")
	config.pruneInfo = readSection(v, config.configFilename, "prune")
	config.checkInfo = readSection(v, config.configFilename, "check")

	// Validate, set defaults
	if len(config.backupInfo) == 0 {
		err = errors.New("no storage locations defined in configuration")
		logError(nil, fmt.Sprint("Error: ", err))
	} else {
		for i, bi := range config.backupInfo {
			if bi["name"] == "" {
				err = fmt.Errorf("missing mandatory storage field: %d.name", i)
				logError(nil, fmt.Sprint("Error: ", err))
			}
		}
	}

	for i, ci := range config.copyInfo {
		if ci["from"] == "" {
			err = fmt.Errorf("missing mandatory from field: %d.from", i)
			logError(nil, fmt.Sprint("Error: ", err))
		}
		if ci["to"] == "" {
			err = fmt.Errorf("missing mandatory to field: %d.to", i)
			logError(nil, fmt.Sprint("Error: ", err))
		}
	}

	if len(config.pruneInfo) == 0 {
		err = errors.New("no prune locations defined in configuration")
		logError(nil, fmt.Sprint("Error: ", err))
	}

	for i, pi := range config.pruneInfo {
		if pi["storage"] == "" {
			err = fmt.Errorf("missing mandatory prune field: %d.storage", i)
			logError(nil, fmt.Sprint("Error: ", err))
		}
		if pi["keep"] == "" {
			err = fmt.Errorf("missing mandatory prune field: %d.keep", i)
			logError(nil, fmt.Sprint("Error: ", err))
		} else {
			// Split/join to get "-keep " before each element
			splitList := strings.Split(pi["keep"], " ")
			for i, element := range splitList {
				splitList[i] = "-keep " + element
			}

			pi["keep"] = strings.Join(splitList, " ")
		}
	}

	if len(config.checkInfo) == 0 {
		err = errors.New("no check locations defined in configuration")
		logError(nil, fmt.Sprint("Error: ", err))
	} else {
		for i, ci := range config.checkInfo {
			if ci["storage"] == "" {
				err = fmt.Errorf("missing mandatory check field: %d.storage", i)
				logError(nil, fmt.Sprint("Error: ", err))
			}
		}
	}
	if err != nil {
		return err
	}
	channels := new(notificationChannels)
	if err := v.UnmarshalKey("notifications", &channels); err != nil {
		return err
	}

	// Load configuration notifications
	// Configure notifiers for onStart notification
	backupOnStartNotifiers, err := configureNotificationChannel(v, channels.OnStart, OnStart)
	if err != nil {
		return err
	}
	// Configure notifiers for onSkip notification
	backupOnSkipNotifiers, err := configureNotificationChannel(v, channels.OnSkip, OnSkipped)
	if err != nil {
		return err
	}

	// Configure notifiers for onSuccess notification
	backupOnSuccessNotifiers, err := configureNotificationChannel(v, channels.OnSuccess, OnSuccess)
	if err != nil {
		return err
	}

	// Configure notifiers for onFailure notification
	backupOnFailureNotifiers, err := configureNotificationChannel(v, channels.OnFailure, OnFailure)
	if err != nil {
		return err
	}
	onStartNotifiers = append(onStartNotifiers, backupOnStartNotifiers...)
	onSkipNotifiers = append(onSkipNotifiers, backupOnSkipNotifiers...)
	onSuccessNotifiers = append(onSuccessNotifiers, backupOnSuccessNotifiers...)
	onFailureNotifiers = append(onFailureNotifiers, backupOnFailureNotifiers...)
	// Generate verbose/debug output if requested (assuming no fatal errors)

	logMessage(nil, fmt.Sprint("Using config file:   ", v.ConfigFileUsed()))

	if verboseFlag {
		logMessage(nil, "")
		logMessage(nil, "Backup Information:")
		logMessage(nil, fmt.Sprintf("  Num\t%-20s%s", "Storage", "Threads"))
		for i := range config.backupInfo {
			var localThreads string
			if _, ok := config.backupInfo[i]["threads"]; ok {
				localThreads = config.backupInfo[i]["threads"]
			}
			logMessage(nil, fmt.Sprintf("  %2d\t%-20s   %-2s", i+1, config.backupInfo[i]["name"], localThreads))
		}
		if len(config.copyInfo) != 0 {
			logMessage(nil, "Copy Information:")
			logMessage(nil, fmt.Sprintf("  Num\t%-20s%-20s%s", "From", "To", "Threads"))
			for i := range config.copyInfo {
				var localThreads string
				if _, ok := config.copyInfo[i]["threads"]; ok {
					localThreads = config.copyInfo[i]["threads"]
				}
				logMessage(nil, fmt.Sprintf("  %2d\t%-20s%-20s   %-2s", i+1, config.copyInfo[i]["from"], config.copyInfo[i]["to"], localThreads))
			}
		}
		logMessage(nil, "")

		logMessage(nil, "Prune Information:")
		for i := range config.pruneInfo {
			logMessage(nil, fmt.Sprintf("  %2d: Storage %s\n      Keep: %s", i+1, config.pruneInfo[i]["storage"], config.pruneInfo[i]["keep"]))
		}
		logMessage(nil, "")

		logMessage(nil, "Check Information:")
		logMessage(nil, fmt.Sprintf("  Num\t%-20s%s", "Storage", "All Snapshots"))
		for i := range config.checkInfo {
			var checkAll string
			if _, ok := config.checkInfo[i]["all"]; ok {
				checkAll = "true"
			}
			logMessage(nil, fmt.Sprintf("  %2d\t%-20s    %-2s", i+1, config.checkInfo[i]["storage"], checkAll))
		}
		logMessage(nil, "")
	}

	if debugFlag {
		logMessage(nil, "")
		logMessage(nil, fmt.Sprint("Backup Info: ", config.backupInfo))
		logMessage(nil, fmt.Sprint("Copy Info: ", config.copyInfo))
		logMessage(nil, fmt.Sprint("Prune Info: ", config.pruneInfo))
		logMessage(nil, fmt.Sprint("Check Info", config.checkInfo))
	}

	return nil
}

func readSection(viper *viper.Viper, filename string, sectionKey string) []map[string]string {
	if viper.IsSet(sectionKey) {
		section := make([]interface{}, 0)
		if viper.IsSet(sectionKey + ".1") {
			// Have we issued our global warning before; if not, do so
			if !oldBackupFileFormat {
				oldBackupFileFormat = true
				logError(nil, "WARNING: Upgrade format of backup configuration "+filename+" to new format!")
			}

			// Keyed by number
			for i := 1; ; i++ {
				key := sectionKey + "." + strconv.Itoa(i)
				if viper.IsSet(key) {
					section = append(section, viper.Get(key))
				} else {
					break
				}
			}
		} else {
			// Array
			section = viper.Get(sectionKey).([]interface{})
		}

		return coerceToArrayOfMapStringString(section[0:])
	}
	return nil
}

func coerceToArrayOfMapStringString(slice []interface{}) []map[string]string {
	section := make([]map[string]string, len(slice))
	for i, item := range slice {
		itemMap := make(map[string]string)
		switch typeList := item.(type) {
		case map[string]interface{}:
			for key, value := range typeList {
				strKey := key
				strValue := fmt.Sprintf("%v", value)
				itemMap[strKey] = strValue
			}
		case map[interface{}]interface{}:
			for key, value := range typeList {
				strKey := fmt.Sprintf("%v", key)
				strValue := fmt.Sprintf("%v", value)
				itemMap[strKey] = strValue
			}
		}
		section[i] = itemMap
	}
	return section
}
