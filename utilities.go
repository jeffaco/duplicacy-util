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
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
)

// Validate that the parameter passed is a valid directory that exists
func validateDirectory(dir string) bool {
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return true
	}

	return false
}

// Get directory where we store our settings. By default, this would be:
//   $HOME/.duplicacy-util
// However, this can be over-ridden by a variety of factors:
//   1. Setting option on the command line (passed by parameter),
//   2. Resetting the value of $HOME environment variable. We will,
//      by default, search in $HOME/.duplicacy-util
//   3. Looking in actual home directory for directory ".duplicacy-util"
func getStorageDirectory(dir string) (string, error) {
	// Was directory passed in (that would override all)
	if dir != "" {
		if validateDirectory(dir) {
			return dir, nil
		}

		// If a directory was passed by parameter, that really should have existed
		err := errors.New("Storage directory '" + dir + "' does not exist")
		logError(nil, fmt.Sprint("Error: ", err))
		return "", err
	}

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error", err)
		return "", err
	}

	potentialDir := path.Join(home, ".duplicacy-util")
	if validateDirectory(potentialDir) {
		return potentialDir, nil
	}

	err = errors.New("Unable to resolve location for storage directory")
	logError(nil, fmt.Sprint("Error: ", err))
	return "", err
}
