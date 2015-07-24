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

package controller

import (
	"os"
	"testing"

	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"github.com/realglobe-Inc/edo-xrs/app/model"
	"gopkg.in/mgo.v2"
)

func TestMain(m *testing.M) {
	session, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer session.Close()
	if err != nil {
		os.Exit(1)
	}
	model.InitDB(session.DB(miscs.GlobalConfig.MongoDB.DBName))
	miscs.InitConfig(os.Getenv("GOPATH") + "/src/github.com/realglobe-Inc/edo-xrs/conf/app.conf")

	code := m.Run()
	os.Exit(code)
}

var singleStatement01 = `
{
  "actor": {
    "objectType": "Agent",
    "name": "Test Name",
    "account": {
      "homePage": "http:\/\/www.example.com",
      "name": "71394872"
    }
  },
  "verb": {
    "id": "http:\/\/www.adlnet.gov\/XAPIprofile\/ran(travelled_a_distance)",
    "display": {
      "ja-JP": "hashita",
      "en-US": "ran"
    }
  },
  "object": {
    "objectType": "StatementRef",
    "id": "1cabcb4f-c41c-49a5-ad89-9a9c8c5fd20a"
  }
}
`

var singleStatement02 = `
  {
    "actor": {
      "objectType": "Agent",
      "name": "object of activity",
      "account": {
        "homePage": "http:\/\/www.example.com\/user\/hoge",
        "name": "71394872"
      }
    },
    "verb": {
      "id": "http:\/\/example.com\/visited",
      "display": {
        "en-US": "will visit"
      }
    },
    "object": {
      "objectType": "Activity",
      "id": "http:\/\/example.com\/website",
      "definition": {
        "name": {
          "en-US": "Some Awsome Website"
        }
      }
    }
  }
`

var singleStatement03 = `
{
  "actor": {
    "objectType": "Agent",
    "name": "Test Name",
    "account": {
      "homePage": "http:\/\/www.example.com",
      "name": "71394872"
    }
  },
  "verb": {
    "id": "http:\/\/www.adlnet.gov\/XAPIprofile\/ran(travelled_a_distance)",
    "display": {
      "ja-JP": "hashita",
      "en-US": "ran"
    }
  },
  "object": {
    "objectType": "StatementRef",
    "id": "1cabcb4f-c41c-49a5-ad89-9a9c8c5fd20a"
  },
  "timestamp": "2013-05-18T05:32:34.804Z"
}
`
