# Copyright © 2018 Jeff Coffler <jeff@taltos.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Pick up the version number
include ./version

COMMIT_SHA=$(shell git rev-parse --short HEAD)
LD_FLAGS="-X main.versionText=${VERSION} -X main.gitHash=${COMMIT_SHA}"

.PHONY: all
## all: Build and test the application
all: build test

.PHONY: build
## build: Build the application
build:
	@echo "Version is: $(VERSION)"
	go build

.PHONY: test
## test: Test the application
test:
	go vet
	go test
	staticcheck
	ineffassign .
	golint -set_exit_status
	gofmt -e -s -w *.go

.PHONY: release
## release: Build release binaries for all platforms
release: clean
	@echo "Version: ${VERSION}, hash: ${COMMIT_SHA}"
	@echo "Building for freebsd x64"
	@GOOS=freebsd GOARCH=amd64	go build -ldflags ${LD_FLAGS} -o duplicacy-util_freebsd_x64_${VERSION}
	@echo "Building for linux arm"
	@GOOS=linux GOARCH=arm 		go build -ldflags ${LD_FLAGS} -o duplicacy-util_linux_arm_${VERSION}
	@echo "Building for linux i386"
	@GOOS=linux GOARCH=386		go build -ldflags ${LD_FLAGS} -o duplicacy-util_linux_i386_${VERSION}
	@echo "Building for linux x64"
	@GOOS=linux GOARCH=amd64	go build -ldflags ${LD_FLAGS} -o duplicacy-util_linux_x64_${VERSION}
	@echo "Building for OS/X x64"
	@GOOS=darwin GOARCH=amd64	go build -ldflags ${LD_FLAGS} -o duplicacy-util_osx_x64_${VERSION}
	@echo "Building for Windows i386"
	@GOOS=windows GOARCH=386	go build -ldflags ${LD_FLAGS} -o duplicacy-util_win_i386_${VERSION}.exe
	@echo "Building for Windows x64"
	@GOOS=windows GOARCH=amd64	go build -ldflags ${LD_FLAGS} -o duplicacy-util_win_x64_${VERSION}.exe

.PHONY: clean
## clean: Clean the binaries / clean source directory
clean:
	@go clean
	@rm -f *~ duplicacy-util duplicacy-util_*_$(VERSION) duplicacy-util_*_$(VERSION).exe

.PHONY: help
## help: Prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
