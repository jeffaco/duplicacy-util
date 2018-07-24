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
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	checkpoint_none = 0 + iota
	checkpoint_backup
	checkpoint_copy
	checkpoint_prune
	checkpoint_check
)

func readCheckpoint() (int, int) {
	// Create new viper instance and read in checkpoint file
	v := viper.New()
	v.AddConfigPath(globalLockDir)
	v.SetConfigName(cmdConfig + "_checkpoint")

	// Read in config file if it is found
	if err := v.ReadInConfig(); err != nil {
		return checkpoint_none, 0
	}

	// Read in checkpoint information
	operation := v.GetInt("Operation")
	iteration := v.GetInt("Iteration")

	// Do some basic validation of checkpoint information
	switch operation {
	case checkpoint_backup:
	case checkpoint_copy:
	case checkpoint_prune:
	case checkpoint_check:
		break

	default:
		operation = checkpoint_none
		iteration = 0
	}

	return operation, iteration
}

func removeCheckpoint() error {
	filename := filepath.Join(globalLockDir, cmdConfig + "_checkpoint.yaml")
	err := os.Remove(filename)
	return err
}

func writeCheckpoint(checkpoint int, iteration int) error {
	filename := filepath.Join(globalLockDir, cmdConfig + "_checkpoint.yaml")

	file, err := os.Create(filename)
	if err != nil { return err }

	// Defer cleanup (if we had an error, delete checkpoint file
	defer func() {
		file.Close()
		if err != nil { os.Remove(filename) }
	}()

	// Write out checkpoint information in YAML format
	_, err = file.WriteString(fmt.Sprintf("Operation: %d\nIteration: %d\n", checkpoint, iteration))
	if err != nil { return err }

	return nil
}
