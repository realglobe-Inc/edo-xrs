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

if grep --quiet "edoxrs" /etc/supervisord.conf>/dev/null; then
  exit 0
fi

yum install -y supervisor

mkdir -p ${WORKSPACE}/logs

cat <<EOF >> /etc/supervisord.conf

[program:edoxrs]
command=/bin/sh /vagrant/tools/vagrant-scripts/run.sh
user=vagrant
log_stdout=true
log_stderr=true
logfile=${WORKSPACE}/logs/edo-xrs.log
logfile_backups=10
EOF

chkconfig supervisord on
