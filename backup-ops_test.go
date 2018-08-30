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
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

// Set up logging for test purposes
func setupLogging() (*log.Logger, *os.File, error) {
	// Create output log file
	file, err := ioutil.TempFile("", "taltos_log")
	if err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return nil, nil, err
	}
	logger := log.New(file, "", 0 /* log.Ltime */)

	return logger, file, nil
}

// Set up arguments for testing of os/exec calls
func fakeBackupOpsCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestBackupOpsHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestRunDuplicacyBackup(t *testing.T) {
	tests := []struct {
		assetInputFragment string
		resultsFile        string
		backupInfo         []map[string]string
		copyInfo           []map[string]string
	}{
		// Dupicacy Error: Enter Backblaze Account ID:Enter Backblaze Application Key:Failed to load the Backblaze B2 storage at b2://hidden-bucket: Authorization failure
		{
			"account_id.log", "account_id.log_results_backup",
			[]map[string]string{
				{"name": "b2", "threads": "10", "vss": "false"},
			},
			[]map[string]string{},
		},
		// Duplicacy Error: Enter storage password:Failed to read the password: EOF
		{
			"storagepw.log", "storagepw.log_results_backup",
			[]map[string]string{
				{"name": "b2", "threads": "5", "vss": "false"},
			},
			[]map[string]string{},
		},
		// Test of long, very involved backup
		{"taltos.log", "taltos.log_results_backup",
			[]map[string]string{
				{"name": "gcd", "threads": "5", "vss": "false"},
				{"name": "azure-direct", "threads": "10", "vss": "false"},
			},
			[]map[string]string{
				{"from": "gcd", "to": "azure", "threads": "5"},
			},
		},
	}

	for _, test := range tests {
		// Set up logging infrastructure
		logger, file, err := setupLogging()
		if err != nil {
			t.Errorf("unexpected error creating log file, got %#v", err)
		}
		loggingSystemDisplayTime = false
		quietFlag = true
		defer func() {
			file.Close()
			os.Remove(file.Name()) // For debugging, might need to leave log file around for perusal

			loggingSystemDisplayTime = true
			quietFlag = false
		}()

		// Initialize data structures for test
		configFile.backupInfo = test.backupInfo
		configFile.copyInfo = test.copyInfo
		mailBody = nil
		//defer os.Remove(file.Name())

		execCommand = fakeBackupOpsCommand
		defer func() { execCommand = exec.Command }()
		if err := performDuplicacyBackup(logger, []string{"testbackup", test.assetInputFragment}); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// Check results of anon function
		expectedOutputInBytes, err := ioutil.ReadFile(path.Join("test/assets", test.resultsFile))
		if err != nil {
			t.Errorf("unable to read backup results file %s", err)
			return
		}
		expectedOutput := string(expectedOutputInBytes)
		actualOutput := strings.Join(mailBody, "\n") + "\n"
		if actualOutput != expectedOutput {
			t.Errorf("result was incorrect, got\n=====\n%s=====\nexpected\n=====\n%s=====", actualOutput, expectedOutput)
		}
	}

	/*
		// Set up logging infrastructure
		logger, file, err := setupLogging()
		if err != nil {
			t.Errorf("unexpected error creating log file, got %#v", err)
		}
		loggingSystemDisplayTime = false
		quietFlag = true
		defer func() {
			file.Close()
			os.Remove(file.Name())	// For debugging, might need to leave log file around for perusal

			loggingSystemDisplayTime = true
			quietFlag = false
		}()

		// Initialize data structures for test
		configFile.backupInfo = []map[string]string {
			{"name": "gcd", "threads": "5", "vss": "false"},
			{"name": "azure-direct", "threads": "10", "vss": "false"},
		}
		configFile.copyInfo = nil
		mailBody = nil
		//defer os.Remove(file.Name())

		execCommand = fakeBackupOpsCommand
		defer func(){ execCommand = exec.Command }()
		if err := performDuplicacyBackup(logger, []string {"testbackup", "taltos.log"}); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// Check results of anon function
		resultsFile := "test/assets/taltos.log_results_backup"
		expectedOutputInBytes, err := ioutil.ReadFile(resultsFile)
		if err != nil {
			t.Errorf("unable to read backup results file%s", err)
			return
		}
		expectedOutput := string(expectedOutputInBytes)
		actualOutput := strings.Join(mailBody, "\n") + "\n"
		if actualOutput != expectedOutput {
			t.Errorf("result was incorrect, got\n=====\n%s=====\nexpected\n=====\n%s=====", actualOutput, expectedOutput)
		}
	*/
}

// Read a file, dumping to stdout. Helper function for TestBackupOpsHelperProcess
func readFileToStdout(logFile string) error {
	file, err := os.OpenFile(logFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// TestBackupOpsHelperProcess isn't a real test; it's a helper process for TestRunDuplicacy*
func TestBackupOpsHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	if cmd != "" {
		// For test, we don't pass a command. If one is found, just return a failure.
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}

	switch args[0] {
	case "testbackup":
		backupFile := args[1]
		args = args[2:]
		backupFile = path.Join("test/assets", backupFile)
		fmt.Fprintf(os.Stdout, "Processing backup file: %q\n", backupFile)
		if err := readFileToStdout(backupFile); err != nil {
			fmt.Printf("error opening assets file: %s\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown argument %q\n", args)
		os.Exit(2)
	}

	os.Exit(0)
}
