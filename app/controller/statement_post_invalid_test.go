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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-martini/martini"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"gopkg.in/mgo.v2"
)

var postInvalidStatementTestCases = []string{
	postInvalidStatement01,
	postInvalidStatement02,
	postInvalidStatement03,
}

func TestPostInvalidStatement(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Post("/:user/:app/statements", c.StoreMultStatement)

	for _, stmt := range postInvalidStatementTestCases {
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/test/statements", strings.NewReader(stmt))
		req.Header.Add("X-Experience-API-Version", "1.0.2")

		m.ServeHTTP(resp, req)

		if got, expected := resp.Code, http.StatusBadRequest; got != expected {
			t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
		}
	}
}

var postInvalidStatement01 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "Statement with multiple IFI.",
      "account": {
        "homePage": "http:\/\/www.example.com",
        "name": "71394872"
      },
      "mbox": "mailto:user@example.com"
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
    "version": "1.0.2"
  }
]
`

var postInvalidStatement02 = `
{ invalid: "json" }
`
var postInvalidStatement03 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "Invalid voided statement (object is not StatementRef)",
      "account": {
        "homePage": "http:\/\/www.example.com",
        "name": "71394872"
      }
    },
    "verb": {
      "id": "http://adlnet.gov/expapi/verbs/voided",
      "display": {
        "en-US": "voided"
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
]
`
