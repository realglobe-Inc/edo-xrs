// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package miscs

import (
	"log"

	"code.google.com/p/gcfg"
)

// Config represent a structure of global config
type Config struct {
	Global struct {
		HostName          string
		MaxStatements     int
		VoidedStatementID string
	}
	MongoDB struct {
		URL    string
		DBName string
	}
	Quota struct {
		UserMaxUsage int64
	}
}

// GlobalConfig is entity of global config
var GlobalConfig Config

func InitConfig(filename string) {
	if err := gcfg.ReadFileInto(&GlobalConfig, filename); err != nil {
		log.Panic(err)
	}
}
