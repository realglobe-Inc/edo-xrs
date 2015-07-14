#!/bin/sh

# Copyright 2015 realglobe, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export WORKSPACE="/home/vagrant"
export GOROOT="/usr/local/go"
export PATH="/usr/local/bin:$PATH:$GOROOT/bin"
export GOPATH="${WORKSPACE}/gopath"

export CPATH="/usr/local/include"
export LIBRARY_PATH="/usr/local/lib"
export LD_LIBRARY_PATH="/usr/local/lib"

mkdir -p "${GOPATH}"
cd "${GOPATH}"

mkdir -p "${WORKSPACE}/logs"
mkdir -p "${GOPATH}/src/github.com/realglobe-Inc/"
if [ ! -f "${GOPATH}/src/github.com/realglobe-Inc/edo-xrs" ]; then
  ln -fs /vagrant ${GOPATH}/src/github.com/realglobe-Inc/edo-xrs
fi

go get github.com/realglobe-Inc/edo-xrs
go install github.com/realglobe-Inc/edo-xrs

#ln -fs /vagrant/jsonschema ${GOPATH}/
#ln -fs /vagrant/conf ${GOPATH}/

./bin/edo-xrs
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
