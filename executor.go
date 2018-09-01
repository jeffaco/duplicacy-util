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
	"bufio"
	"os/exec"
)

var execCommand = exec.Command

func executorStdout(cmdName string, cmdArgs []string) (stdOut []byte, err error) {
	stdOut, err = exec.Command(cmdName, cmdArgs...).Output()
	return stdOut, err
}

func executor(cmdName string, cmdArgs []string, defDir string, output func(string)) error {
	cmd := execCommand(cmdName, cmdArgs...)
	cmd.Dir = defDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		output(scanner.Text())
	}

	err = cmd.Wait()
	return err
}
