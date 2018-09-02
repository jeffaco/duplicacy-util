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
	"os"
	"path"
	"testing"
)

func TestGetStorageDirectoryDoesNotExist(t *testing.T) {
	// Suspend error output
	quietFlag = true
	defer func() { quietFlag = false }()

	_, err := getStorageDirectory(path.Join(os.TempDir(), "ThisIsADirectoryThatShouldNotExist_"+randomStringBytes(6)))
	if err == nil {
		t.Errorf("successful call to getStorageDirectory with non-existent directory")
	}

	return
}

func TestGetStorageDirectoryWithHomeDirectory(t *testing.T) {
	// The actual home directory should always exist
	_, err := getStorageDirectory(os.Getenv("HOME"))
	if err != nil {
		t.Errorf("error result from getStorageDirectory with home directory: %s", err)
	}

	return
}

func TestGetStorageDirectoryWithDefaults(t *testing.T) {
	defaultLocation := path.Join(os.Getenv("HOME"), ".duplicacy-util")
	if !validateDirectory(defaultLocation) {
		err := os.Mkdir(defaultLocation, os.ModePerm)
		if err != nil {
			t.Errorf("error creating directory: %s", err)
			return
		}
		defer os.Remove(defaultLocation)
	}

	_, err := getStorageDirectory("")
	if err != nil {
		t.Errorf("failed call to getStorageDirectory with directory %s", defaultLocation)
	}
}

func TestGetStorageDirectoryWithDirectory(t *testing.T) {
	temporaryFile := path.Join(os.TempDir(), "duplicacy-util-"+randomStringBytes(6))
	err := os.Mkdir(temporaryFile, os.ModePerm)
	if err != nil {
		t.Errorf("failed to create temporary directory %s: %s", temporaryFile, err)
		return
	}
	defer os.Remove(temporaryFile)

	_, err = getStorageDirectory(temporaryFile)
	if err != nil {
		t.Errorf("failed call to getStorageDirectory with directory %s: %s", temporaryFile, err)
	}
}

func TestGetStorageDirectoryWithFileInsteadOfDirectory(t *testing.T) {
	// Suspend error output
	quietFlag = true
	defer func() { quietFlag = false }()

	temporaryFile := path.Join(os.TempDir(), "duplicacy-util-"+randomStringBytes(6))
	file, err := os.Create(temporaryFile)
	if err != nil {
		t.Errorf("failed to create temporary file %s: %s", temporaryFile, err)
		return
	}
	defer os.Remove(temporaryFile)
	defer file.Close()

	_, err = getStorageDirectory(temporaryFile)
	if err == nil {
		t.Errorf("successful call to getStorageDirectory with file %s: %s", temporaryFile, err)
	}
}
