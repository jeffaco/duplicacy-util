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

.PHONY: all build test release clean

# Pick up the version number
include ./version

all: build test

build:
	@echo "Version is: $(VERSION)"
	go build

test:
	go vet
	go test
	staticcheck
	ineffassign .
	golint -set_exit_status
	gofmt -e -s -w *.go

release:
	@./build.sh

clean:
	rm -f *~ duplicacy-util duplicacy-util_*_$(VERSION) duplicacy-util_*_$(VERSION).exe
