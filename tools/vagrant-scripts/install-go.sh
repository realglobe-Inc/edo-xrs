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

if [ -d "/usr/local/go" ]; then
  exit 0
fi

cd "/tmp"
wget --quiet "https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz"
tar -C "/usr/local" -xzf "go1.4.2.linux-amd64.tar.gz"
