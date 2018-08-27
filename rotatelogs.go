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
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	// Look up file times in platform independent way
	"github.com/djherbis/times"
)

func rotateLogFiles() error {
	logFileRoot := filepath.Join(globalLogDir, cmdConfig) + ".log"

	// Kick the older log files up by a count of one
	for i := globalLogFileCount - 2; i >= 1; i-- {
		// We don't check for existence of files, so ignore any errors
		os.Rename(logFileRoot+"."+strconv.Itoa(i)+".gz",
			logFileRoot+"."+strconv.Itoa(i+1)+".gz")
	}

	// If uncompressed log file exists, rename it and compress it
	if _, err := os.Stat(logFileRoot); os.IsNotExist(err) {
		return nil
	}

	//
	// Compress <file.log> into <file.log.1.gz>
	//

	// We want to save original time, so look up old time before compression
	t, err := times.Stat(logFileRoot)
	if err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}

	atime := t.AccessTime()
	mtime := t.ModTime()

	// Now compress the file
	if err = compressLogFile(logFileRoot); err != nil {
		return err
	}

	// Restore the file times
	if err = os.Chtimes(logFileRoot+".1.gz", atime, mtime); err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}

	return nil
}

func compressLogFile(logFileRoot string) error {
	reader, err := os.Open(logFileRoot)
	if err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}

	writer, err := os.Create(logFileRoot + ".1.gz")
	if err != nil {
		reader.Close()
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = logFileRoot + ".1.gz"
	defer archiver.Close()

	if _, err := io.Copy(archiver, reader); err != nil {
		logError(nil, fmt.Sprint("Error: ", err))
		return err
	}

	return nil
}
