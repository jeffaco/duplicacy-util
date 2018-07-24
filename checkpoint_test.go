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
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var randSrc = rand.NewSource(time.Now().UnixNano())

func randomStringBytes(n int) string {
	// Generate a string of bytes as recommended in Stack Overflow:
	//
	// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, randSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func TestCheckPointRead_NotFound(t *testing.T) {
	globalLockDir = os.TempDir()
	cmdConfig = "SomeCheckpointFileThatShouldNotExist"
	checkpoint, iteration := readCheckpoint()

	if checkpoint != checkpoint_none {
		t.Errorf("Checkpoint is incorrect, got '%d', expected '%d'.", checkpoint, checkpoint_none)
	}
	if iteration != 0 {
		t.Errorf("iteration is incorrect, got '%d', expected '%d'.", iteration, 0)
	}
}

func TestCheckpointRead_KnownGoodFile(t *testing.T) {
	// Use random name for the checkpoint file
	globalLockDir = os.TempDir()
	cmdConfig = "checkpoint-file-" + randomStringBytes(6)
	filename := path.Join(path.Join(globalLockDir, cmdConfig + "_checkpoint.yaml"))

	// Write out the checkpoint file ourselves (isolate testing to readCheckpoint())
	file, err := os.Create(filename)
	if err != nil { t.Errorf("Error creating YAML file: %s: %s.", filename, err) }
	defer func() {
		file.Close()
		os.Remove(filename)
	}()

	_, err = file.WriteString(fmt.Sprintf("Operation: %d\nIteration: %d\n", checkpoint_prune, 42))
	if err != nil { t.Errorf("Error writing to checkpoint file: %s.", err) }
	file.Close()

	// Now test reading the checkpoint file
	checkpoint, iteration := readCheckpoint()
	if checkpoint != checkpoint_prune {
		t.Errorf("Incorrect checkpoint read, got '%d', expected '%d'.", checkpoint, checkpoint_prune)
	}
	if iteration != 42 {
		t.Errorf("Incorrect iteration read, got '%d', expected '%d'.", iteration, 42)
	}
}

func TestCheckpointRead_KnownBadFile(t *testing.T) {
	// Use random name for the checkpoint file
	globalLockDir = os.TempDir()
	cmdConfig = "checkpoint-file-" + randomStringBytes(6)
	filename := path.Join(path.Join(globalLockDir, cmdConfig + "_checkpoint.yaml"))

	// Write out the checkpoint file ourselves (isolate testing to readCheckpoint())
	file, err := os.Create(filename)
	if err != nil { t.Errorf("Error creating YAML file: %s: %s.", filename, err) }
	defer func() {
		file.Close()
		os.Remove(filename)
	}()

	// Write out a known-invalid checkpoint file
	_, err = file.WriteString(fmt.Sprintf("Operation: %d\nIteration: %d\n", 42, 84))
	if err != nil { t.Errorf("Error writing to checkpoint file: %s.", err) }
	file.Close()

	// Now test reading the checkpoint file
	checkpoint, iteration := readCheckpoint()
	if checkpoint != checkpoint_none {
		t.Errorf("Incorrect checkpoint read, got '%d', expected '%d'.", checkpoint, checkpoint_none)
	}
	if iteration != 0 {
		t.Errorf("Incorrect iteration read, got '%d', expected '%d'.", iteration, 0)
	}
}

func TestCheckpointWrite(t *testing.T) {
	testEntries := []struct {
		checkpointExpected int
		iterationExpected int
		valid bool
	}{
		{checkpoint_backup, 10, true},
		{checkpoint_copy, 2, true},
		{checkpoint_prune, 1, true},
		{checkpoint_check, 3, true},

		{10, 5, false},
		{checkpoint_check + 1, 6, false},
		{-1, 7, false},
	}

	for _, entry := range testEntries {
		// Use random name for the checkpoint file
		globalLockDir = os.TempDir()
		cmdConfig = "checkpoint-file-" + randomStringBytes(6)

		// Write the checkpoint file out, re-read it
		writeCheckpoint(entry.checkpointExpected, entry.iterationExpected)
		defer removeCheckpoint()

		checkpoint, iteration := readCheckpoint()

		// If this is an invalid entry, change what we expect
		if ! entry.valid { entry.checkpointExpected = 0; entry.iterationExpected = 0 }

		if checkpoint != entry.checkpointExpected {
			t.Errorf("Incorrect checkpoint read, got '%d', expected '%d'.", checkpoint, entry.checkpointExpected)
		}
		if iteration != entry.iterationExpected {
			t.Errorf("Incorrect iteration read, got '%d', expected '%d'.", iteration, entry.iterationExpected)
		}

		if err := removeCheckpoint(); err != nil {
			t.Errorf("Error removing checkpoint file: %s", err)
		}
	}
}
