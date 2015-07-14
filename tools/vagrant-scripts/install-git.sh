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

# CentOS 6 リポジトリにある git は古いため、go get した際にエラーが発生する
# このことから、git は yum でインストールするのではなく、ソースからビルドし
# インストールを行う

yum install -y gcc zlib zlib-devel perl-ExtUtils-MakeMaker
yum install -y curl-devel expat-devel gettext-devel openssl-devel

export PATH="$PATH:/usr/local/bin"

if which git>/dev/null; then
  exit 0
fi

cd /tmp
wget --quiet "https://www.kernel.org/pub/software/scm/git/git-2.4.4.tar.gz"
tar xzf "git-2.4.4.tar.gz"

cd "git-2.4.4"
./configure --prefix="/usr/local"
make
make install

