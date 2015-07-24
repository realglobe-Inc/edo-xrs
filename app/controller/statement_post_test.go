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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/acceptlang"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"gopkg.in/mgo.v2"
)

var postStatementTestCases = []string{
	postStatement01,
	postStatement02,
	postStatement03,
	postStatement04,
	postStatement05,
	postStatement06,
	postStatement07,
	postStatement08,
	postStatement09,
	postStatement10,
	postStatement11,
	postStatement12,
	postStatement13,
	postStatement14,
	postStatement15,
	postStatement16,
	postStatement17,
	postStatement18,
	postStatement19,
	postStatement20,
}

func TestPostStatement(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Post("/:user/:app/statements", c.StoreMultStatement)

	for _, stmt := range postStatementTestCases {
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/test/statements", strings.NewReader(stmt))
		req.Header.Add("X-Experience-API-Version", "1.0.2")

		m.ServeHTTP(resp, req)

		if got, expected := resp.Code, http.StatusOK; got != expected {
			t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
		}
	}
}

func TestPostAndGetStatement(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Get("/:user/:app/statements", acceptlang.Languages(), c.FindStatement)
	m.Post("/:user/:app/statements", c.StoreMultStatement)

	for _, stmt := range postStatementTestCases {
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/test/statements", strings.NewReader(stmt))
		req.Header.Add("X-Experience-API-Version", "1.0.2")

		m.ServeHTTP(resp, req)

		if got, expected := resp.Code, http.StatusOK; got != expected {
			t.Fatalf("Expected %v response code from post single statement; got %d", expected, got)
		}

		var statementIDs []interface{}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("A POST request with statement(s) returns empty body.")
		}

		err = json.Unmarshal(body, &statementIDs)
		if err != nil {
			t.Fatalf("A POST request with statement(s) returns non-json string.")
		}

		for _, v := range statementIDs {
			id := v.(string)

			resp2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", "/test/test/statements?statementId="+id, nil)
			req2.Header.Add("X-Experience-API-Version", "1.0.2")

			m.ServeHTTP(resp2, req2)

			if got, expected := resp2.Code, http.StatusOK; got != expected {
				t.Fatalf("Expected %v response code from get single statement; got %d", expected, got)
			}
		}
	}
}

var postStatement01 = `
[
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
]
`

var postStatement02 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "Verb with single display unit.",
      "account": {
        "homePage": "http:\/\/www.example.com",
        "name": "71394872"
      }
    },
    "verb": {
      "id": "http:\/\/www.adlnet.gov\/XAPIprofile\/ran(travelled_a_distance)",
      "display": {
        "en-US": "ran"
      }
    },
    "object": {
      "objectType": "StatementRef",
      "id": "1cabcb4f-c41c-49a5-ad89-9a9c8c5fd20a"
    }
  }
]
`
var postStatement03 = `
{
  "actor": {
    "objectType": "Agent",
    "name": "Project Tin Can API",
    "mbox": "mailto:user@example.com"
  },
  "verb": {
    "id": "http:\/\/adnet.gov\/expapi\/verbs\/created",
    "display": {
      "en-US": "created"
    }
  },
  "object": {
    "id": "http:\/\/example.adlnet.gov\/xapi\/example\/statement",
    "definition": {
      "name": {
        "en-US": "simple statement"
      },
      "description": {
        "en-US": "A simple Experience API statement. Note that the xRS does not need to have any prior information about the Actor (learner), the verb, ro the Activity\/object."
      }
    }
  }
}
`

var postStatement04 = `
{
  "actor": {
    "objectType": "Agent",
    "name": "Example Learner",
    "mbox": "mailto:example.lerner@adlnet.gov"
  },
  "verb": {
    "id": "http:\/\/adlnet.gov\/expapi\/verbs\/attempted",
    "display": {
      "en-US": "attempted"
    }
  },
  "object": {
    "id": "http:\/\/example.adlnet.gov\/xampi\/example\/simpleCBT",
    "definition": {
      "name": {
        "en-US": "simple CBT course"
      },
      "description": {
        "en-US": "A fictitious example CBT course."
      }
    }
  },
  "result": {
    "score": {
      "scaled": 0.95
    },
    "success": true,
    "completion": true
  }
}
`

var postStatement05 = `
{
  "actor": {
    "name": "Team PB",
    "mbox": "mailto:teampb@example.com",
    "member": [
      {
        "name": "Andrew Downes",
        "account": {
          "homePage": "http:\/\/www.example.com",
          "name": "13936749"
        },
        "objectType": "Agent"
      },
      {
        "name": "Toby Nichols",
        "openid": "http:\/\/toby.openid.example.org\/",
        "objectType": "Agent"
      },
      {
        "name": "Ena Hills",
        "mbox_sha1sum": "ebd31e95054c018b10727ccffd2ef2ec3a016ee9",
        "objectType": "Agent"
      }
    ],
    "objectType": "Group"
  },
  "verb": {
    "id": "http:\/\/adlnet.gov\/expapi\/verbs\/attended",
    "display": {
      "en-GB": "attended",
      "en-US": "attended"
    }
  },
  "result": {
    "extensions": {
      "http://example.com/profiles/meetings/resultextensions/minuteslocation": "X:\\meetings\\minutes\\examplemeeting.one"
    },
    "success": true,
    "completion": true,
    "response": "We agreed on some example actions.",
    "duration": "PT1H0M0S"
  },
  "context": {
    "registration": "ec531277-b57b-4c15-8d91-d292c5b2b8f7",
    "contextActivities": {
      "parent": [
        {
          "id": "http:\/\/www.example.com\/meetings\/series\/267",
          "objectType": "Activity"
        }
      ],
      "category": [
        {
          "id": "http:\/\/www.example.com\/meetings\/categories\/teammeeting",
          "objectType": "Activity",
          "definition": {
            "name": {
              "en": "team meeting"
            },
            "description": {
              "en": "A category of meeting used for regular team meetings."
            },
            "type": "http:\/\/example.com\/expapi\/activities\/meetingcategory"
          }
        }
      ],
      "other": [
        {
          "id": "http:\/\/www.example.com\/meetings\/occurances\/34257",
          "objectType": "Activity"
        },
        {
          "id": "http:\/\/www.example.com\/meetings\/occurances\/3425567",
          "objectType": "Activity"
        }
      ]
    },
    "instructor": {
      "name": "Andrew Downes",
      "account": {
        "homePage": "http:\/\/www.example.com",
        "name": "13936749"
      },
      "objectType": "Agent"
    },
    "team": {
      "name": "Team PB",
      "mbox": "mailto:teampb@example.com",
      "objectType": "Group"
    },
    "platform": "Example virtual meeting software",
    "language": "tlh",
    "statement": {
      "objectType": "StatementRef",
      "id": "6690e6c9-3ef0-4ed3-8b37-7f3964730bee"
    }
  },
  "timestamp": "2013-05-18T05:32:34.804Z",
  "authority": {
    "account": {
      "homePage": "http:\/\/cloud.scorm.com\/",
      "name": "anonymous"
    },
    "objectType": "Agent"
  },
  "version": "1.0.2",
  "object": {
    "id": "http:\/\/www.example.com\/meetings\/occurances\/34534",
    "definition": {
      "extensions": {
        "http://example.com/profiles/meetings/activitydefinitionextensions/room": {
          "name": "Kilby",
          "id": "http:\/\/example.com\/rooms\/342"
        }
      },
      "name": {
        "en-GB": "example meeting",
        "en-US": "example meeting"
      },
      "description": {
        "en-GB": "An example meeting that happened on a specific occasion with certain people present.",
        "en-US": "An example meeting that happened on a specific occasion with certain people present."
      },
      "type": "http:\/\/adlnet.gov\/expapi\/activities\/meeting",
      "moreInfo": "http:\/\/virtualmeeting.example.com\/345256"
    },
    "objectType": "Activity"
  }
}
`
var postStatement06 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "Verb without display.",
      "account": {
        "homePage": "http:\/\/www.example.com",
        "name": "71394872"
      }
    },
    "verb": {
      "id": "http:\/\/www.adlnet.gov\/XAPIprofile\/ran(travelled_a_distance)"
    },
    "object": {
      "objectType": "StatementRef",
      "id": "1cabcb4f-c41c-49a5-ad89-9a9c8c5fd20a"
    }
  }
]
`

var postStatement07 = `
[
  {
    "actor": {
      "objectType": "Group",
      "member": [
        {
          "objectType": "Agent",
          "account": {
            "homePage": "http:\/\/www.example.com",
            "name": "71394872"
          }
        }
      ],
      "account": {
        "homePage": "http:\/\/www.example.com",
        "name": "7777777"
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
]
`

var postStatement08 = `
[
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
]
`

var postStatement09 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "object with description",
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
        },
	"description": {
	  "en-US": "The website explains the history and the meaninig of example."
	}
      }
    }
  }
]
`
var postStatement10 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "object of Activity with moreInfo",
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
        },
	"moreInfo": "http://www.merriam-webster.com/dictionary/example"
      }
    }
  }
]
`

var postStatement11 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "object of Activity with interaction activities.",
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
]
`

var postStatement12 = `
[
  {
    "actor": {
      "objectType": "Group",
      "name": "multi member group",
      "member": [
        {
          "objectType": "Agent",
          "account": {
            "homePage": "http:\/\/www.example.com",
            "name": "71394874"
          }
        },
        {
          "objectType": "Agent",
          "account": {
            "homePage": "http:\/\/www.example.com",
            "name": "71394875"
          }
        }
      ],
      "account": {
        "homePage": "http:\/\/www.example.com",
        "name": "7777777"
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
]
`

var postStatement13 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "Statement with result",
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
    "result": {
      "score": {
        "scaled": 0.27,
        "raw": 50,
        "min": 0,
        "max": 100
      },
      "success": true,
      "completion": true,
      "response": "OK",
      "duration": "2004-04-01T12:00:00+09:00/2007-08-31T15:00:00+09:00"
    }
  }
]
`

var postStatement14 = `
[
  {
    "actor": {
      "objectType": "Group",
      "name": "anon group",
      "member": [
        {
          "objectType": "Agent",
          "account": {
            "homePage": "http:\/\/www.example.com",
            "name": "71394874"
          }
        },
        {
          "objectType": "Agent",
          "account": {
            "homePage": "http:\/\/www.example.com",
            "name": "71394875"
          }
        }
      ]
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
]
`

var postStatement15 = `
[
  {
    "actor": {
      "objectType": "Group",
      "name": "Goup without member",
      "account": {
	"homePage": "http://www.example.com/",
	"name": "7777777"
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
]
`

var postStatement16 = `
[
  {
    "actor": {
      "objectType": "Agent",
      "name": "Statement with timestamp",
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
    "timestamp": "2008-09-08T22:47:31-07:00"
  }
]
`

var postStatement17 = `
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

var postStatement18 = `
[
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
    "authority": {
      "objectType": "Group",
      "member": [
	{
	  "account": {
	    "homePage": "http://example.com/xAPI/OAuth/Token",
	    "name": "oauth_consumer_x75b"
	  }
	},
	{
	  "mbox": "mailto:bob@example.com"
	}
      ]
    }
  }
]
`

var postStatement19 = `
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

var postStatement20 = `
[
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
    "version": "1.0.2"
  }
]
`
