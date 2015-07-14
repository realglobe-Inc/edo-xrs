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

if [ -f "/etc/yum.repos.d/mongodb-org-stable.repo" ]; then
  exit 0
fi

cat <<EOF > /etc/yum.repos.d/mongodb-org-stable.repo
[mongodb-org-3.0]
name=MongoDB Repository
baseurl=http://repo.mongodb.org/yum/redhat/\$releasever/mongodb-org/stable/x86_64/
gpgcheck=0
enabled=1
EOF

yum update -y
yum install -y mongodb-org

service mongod start
chkconfig mongod on
