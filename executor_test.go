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
// limitations under the License.package utils

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Testing code which invokes os/exec is a little tricky (you need to set up
// the os/exec testing framework for unit tests). We do this in numerous other
// tests. This test exists as a very simplistic bare bones test of code that
// sets up testing framework for os/exec calls.

// Set up arguments for testing of os/exec calls
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecutorHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// Simple test to just pick up and verify expected output
func TestRunExecutor(t *testing.T) {
	// Handling when processing output from generic "duplicacy" command
	outputArray := []string{}
	anon := func(s string) { outputArray = append(outputArray, s) }

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	err := executor(duplicacyPath, []string{"-some", "-fake", "-args"}, configFile.repoDir, anon)
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	// Check results of anon function
	expectedOutput := "This is the expected\noutput\n"
	actualOutput := strings.Join(outputArray, "\n") + "\n"
	if actualOutput != expectedOutput {
		t.Errorf("result was incorrect, got '%s', expected '%s'.", actualOutput, expectedOutput)
	}
}

// TestExecutorHelperProcess isn't a real test; it's a helper process for TestRunExecutor*
func TestExecutorHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// some code here to check arguments perhaps?
	fmt.Fprintf(os.Stdout, "This is the expected\noutput\n")
	os.Exit(0)
}
