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
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/acceptlang"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
)

func initDatabase(t *testing.T) *mgo.Session {
	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	if err != nil {
		t.Fatal(err)
	}
	sess.SetMode(mgo.Strong, true)

	return sess
}

func initHandler(s *mgo.Session) *martini.ClassicMartini {
	mart := martini.Classic()
	hand := New(s)
	mart.Get("/:user/:app/statements", acceptlang.Languages(), hand.FindStatement)
	mart.Put("/:user/:app/statements", hand.StoreStatement)
	mart.Post("/:user/:app/statements", hand.StoreMultStatement)

	return mart
}

func putStatement(t *testing.T, mart *martini.ClassicMartini, stmt, id string) {
	req, _ := http.NewRequest("PUT",
		"/test/test/statements?statementId="+id,
		strings.NewReader(stmt),
	)

	resp := httptest.NewRecorder()
	req.Header.Add("X-Experience-API-Version", "1.0.2")
	mart.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusNoContent; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func getStatementWithHeader(t *testing.T, mart *martini.ClassicMartini, v *url.Values) ([]byte, http.Header) {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test/test/statements?"+v.Encode(), nil)
	fatalIfError(t, err)

	req.Header.Add("X-Experience-API-Version", "1.0.2")
	mart.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusOK; got != expected {
		t.Fatalf("Expected %v response code from get statement(s); got %d", expected, got)
	}

	body, err := ioutil.ReadAll(resp.Body)
	fatalIfError(t, err)

	return body, resp.Header()
}

func getStatement(t *testing.T, mart *martini.ClassicMartini, v *url.Values) []byte {
	body, _ := getStatementWithHeader(t, mart, v)
	return body
}

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetMultStatementWithAscending(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatement01))
	fatalIfError(t, err)

	vsuffix := uuid.NewV4().String()
	verbID := "http://example.com/realglobe/XAPIprofile/test-" + vsuffix
	_, err = stmt.SetP(verbID, "verb.id")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)
	putStatement(t, mart, stmt.String(), id2)

	// construct query
	v := &url.Values{}
	v.Add("verb", verbID)
	v.Add("ascending", "true")

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 2 {
		t.Fatalf("Expected 2 statements in response; got %d", cnt)
	}

	s1, err := respstmt.ArrayElement(1, "statements")
	fatalIfError(t, err)
	if id, ok := s1.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid order of statement array")
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id2 {
		t.Fatalf("Got invalid order of statement array")
	}
}

func TestGetMultStatementWithLimit(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatement01))
	fatalIfError(t, err)

	vsuffix := uuid.NewV4().String()
	verbID := "http://example.com/realglobe/XAPIprofile/test-" + vsuffix
	_, err = stmt.SetP(verbID, "verb.id")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)
	putStatement(t, mart, stmt.String(), id2)

	// construct query
	v := &url.Values{}
	v.Add("verb", verbID)
	v.Add("limit", "1")

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 1 {
		t.Fatalf("Expected 1 statements in response; got %d", cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid order of statement array")
	}
}

var singleStatementWithLangMap = `
{
    "actor": {
        "objectType": "Agent",
        "name": "Test Canonical LangMap",
        "mbox": "mailto:test@example.com"
    },
    "verb": {
        "id": "http://www.adlnet.gov/XAPIprofile/ran(travelled_a_distance)",
        "display": {
            "ja-JP": "hashita",
            "en-US": "ran"
        }
    },
    "object": {
        "objectType": "Activity",
        "id": "http://example.com/XAPIProfile/activity/1cabcb4f-c41c-49a5-ad89-9a9c8c5fd20a",
        "definition": {
            "description": {
                "ja-JP": "Test Activity Description (in Japanese)",
                "en-US": "Test Activity Description"
            },
            "type": "http://adlnet.gov/expapi/activities/cmi.interaction",
            "interactionType": "likert",
            "correctResponsesPattern": [
                "likert_3"
            ],
            "scale": [
                {
                    "id": "likert_0",
                    "description": {
                        "ja-JP": "It's OK (in Japanese)",
                        "en-US": "It's OK"
                    }
                }
            ]
        }
    }
}
`

func TestGetStatementWithCanonicalFormat(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatementWithLangMap))
	fatalIfError(t, err)

	vsuffix := uuid.NewV4().String()
	verbID := "http://example.com/realglobe/XAPIprofile/test-" + vsuffix
	_, err = stmt.SetP(verbID, "verb.id")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)

	// construct query
	v := &url.Values{}
	v.Add("verb", verbID)
	v.Add("format", "canonical")

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 1 {
		t.Fatalf("Expected 1 statements in response; got %d", cnt)
	}

	// check ID
	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid statement")
	}

	// check LangMaps
	scale, err := s0.ArrayElement(0, "object", "definition", "scale")
	fatalIfError(t, err)
	if _, ok := scale.Search("description", "ja-JP").Data().(string); !ok {
		t.Fatalf("ja-JP field not found")
	}
	if _, ok := scale.Search("description", "en-US").Data().(string); ok {
		t.Fatalf("en-US field is in response (not removed)")
	}

	if _, ok := s0.Search("object", "definition", "description", "ja-JP").Data().(string); !ok {
		t.Fatalf("ja-JP field not found")
	}
	if _, ok := s0.Search("object", "definition", "description", "en-US").Data().(string); ok {
		t.Fatalf("en-US field is in response (not removed)")
	}
}

func TestGetMultStatementWithVerbID(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatement01))
	fatalIfError(t, err)

	vsuffix := uuid.NewV4().String()
	verbID := "http://example.com/realglobe/XAPIprofile/test-" + vsuffix
	_, err = stmt.SetP(verbID, "verb.id")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)
	putStatement(t, mart, stmt.String(), id2)

	// construct query
	v := &url.Values{}
	v.Add("verb", verbID)

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 2 {
		t.Fatalf("Expected 2 statements in response; got %d", cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid order of statement array")
	}

	s1, err := respstmt.ArrayElement(1, "statements")
	fatalIfError(t, err)
	if id, ok := s1.Search("id").Data().(string); !ok || id != id2 {
		t.Fatalf("Got invalid order of statement array")
	}
}

func TestGetMultStatementWithSinceUntil(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()
	id3 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatement01))
	fatalIfError(t, err)

	vsuffix := uuid.NewV4().String()
	verbID := "http://example.com/realglobe/XAPIprofile/test-" + vsuffix
	_, err = stmt.SetP(verbID, "verb.id")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)
	since := time.Now()
	time.Sleep(time.Millisecond)
	putStatement(t, mart, stmt.String(), id2)
	time.Sleep(time.Millisecond)
	until := time.Now()
	putStatement(t, mart, stmt.String(), id3)

	// construct query
	v := &url.Values{}
	v.Add("verb", verbID)
	v.Add("since", since.Format(time.RFC3339Nano))
	v.Add("until", until.Format(time.RFC3339Nano))

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 1 {
		t.Fatalf("Expected 1 statements in response; got %d", cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id2 {
		t.Fatalf("Got invalid order of statement array")
	}
}

func TestInvalidXAPIVersion(t *testing.T) {
	mart := initHandler(nil)
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test/test/statements", nil)
	fatalIfError(t, err)
	req.Header.Add("X-Experience-API-Version", "0.0.0")
	mart.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusBadRequest; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}
}

func TestGetMultStatementWithActivity(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatement02))
	fatalIfError(t, err)

	asuffix := uuid.NewV4().String()
	activityID := "http://example.com/realglobe/website/" + asuffix
	_, err = stmt.SetP(activityID, "object.id")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)
	putStatement(t, mart, stmt.String(), id2)

	// construct query
	v := &url.Values{}
	v.Add("activity", activityID)

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 2 {
		t.Fatalf("Expected 2 statements in response; got %d", cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid order of statement array")
	}

	s1, err := respstmt.ArrayElement(1, "statements")
	fatalIfError(t, err)
	if id, ok := s1.Search("id").Data().(string); !ok || id != id2 {
		t.Fatalf("Got invalid order of statement array")
	}
}

func TestGetMultStatementWithRegistration(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatement02))
	fatalIfError(t, err)

	registrationID := uuid.NewV4().String()
	_, err = stmt.SetP(registrationID, "context.registration")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)
	putStatement(t, mart, stmt.String(), id2)

	// construct query
	v := &url.Values{}
	v.Add("registration", registrationID)

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 2 {
		t.Fatalf("Expected 2 statements in response; got %d", cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid order of statement array")
	}

	s1, err := respstmt.ArrayElement(1, "statements")
	fatalIfError(t, err)
	if id, ok := s1.Search("id").Data().(string); !ok || id != id2 {
		t.Fatalf("Got invalid order of statement array")
	}
}

var statetmentAgentBoilerplate = `
{
  "actor": {
    "objectType": "Agent",
    "name": "Test Agent Boilerplate"
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

func appendAccount(t *testing.T, stmt *gabs.Container) *gabs.Container {
	stmt, err := gabs.ParseJSON([]byte(stmt.String()))
	fatalIfError(t, err)

	suffix := uuid.NewV4().String()
	name := "name-example-" + suffix
	homePage := "http://example.com"
	_, err = stmt.SetP(name, "actor.account.name")
	fatalIfError(t, err)
	_, err = stmt.SetP(homePage, "actor.account.homePage")
	fatalIfError(t, err)

	return stmt
}

func appendMbox(t *testing.T, stmt *gabs.Container) *gabs.Container {
	stmt, err := gabs.ParseJSON([]byte(stmt.String()))
	fatalIfError(t, err)

	domainPrefix := uuid.NewV4().String()
	mbox := "mbox:test@" + domainPrefix + ".example.com"
	_, err = stmt.SetP(mbox, "actor.mbox")
	fatalIfError(t, err)

	return stmt
}

func appendMboxSHA1(t *testing.T, stmt *gabs.Container) *gabs.Container {
	stmt, err := gabs.ParseJSON([]byte(stmt.String()))
	fatalIfError(t, err)

	data := make([]byte, 16)
	_, err = rand.Read(data)
	fatalIfError(t, err)
	mboxSHA1 := fmt.Sprintf("%x", sha1.Sum(data))
	_, err = stmt.SetP(mboxSHA1, "actor.mbox_sha1sum")
	fatalIfError(t, err)

	return stmt
}

func appendOpenID(t *testing.T, stmt *gabs.Container) *gabs.Container {
	stmt, err := gabs.ParseJSON([]byte(stmt.String()))
	fatalIfError(t, err)

	domainPrefix := uuid.NewV4().String()
	openID := "http://" + domainPrefix + ".openid.example.com/"
	_, err = stmt.SetP(openID, "actor.openid")
	fatalIfError(t, err)

	return stmt
}

func TestGetMultStatementWithActorAgentAccount(t *testing.T) {
	stmt, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)

	testGetMultStatementOfActor(t, appendAccount(t, stmt))
}

func TestGetMultStatementWithActorAgentMbox(t *testing.T) {
	stmt, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)

	testGetMultStatementOfActor(t, appendMbox(t, stmt))
}

func TestGetMultStatementWithActorAgentMboxSha1(t *testing.T) {
	stmt, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)

	testGetMultStatementOfActor(t, appendMboxSHA1(t, stmt))
}

func TestGetMultStatementWithActorAgentOpenID(t *testing.T) {
	stmt, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)

	testGetMultStatementOfActor(t, appendOpenID(t, stmt))
}

func testGetMultStatementOfActor(t *testing.T, stmt *gabs.Container) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()

	putStatement(t, mart, stmt.String(), id1)
	putStatement(t, mart, stmt.String(), id2)

	// construct query
	v := &url.Values{}
	v.Add("agent", stmt.Search("actor").String())

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 2 {
		t.Fatalf("Expected 2 statements in response; got %d", cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid order of statement array")
	}

	s1, err := respstmt.ArrayElement(1, "statements")
	fatalIfError(t, err)
	if id, ok := s1.Search("id").Data().(string); !ok || id != id2 {
		t.Fatalf("Got invalid order of statement array")
	}
}

func TestGetMultStatementWithActorGroupAccount(t *testing.T) {
	group, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)
	_, err = group.SetP("Group", "actor.objectType")
	fatalIfError(t, err)
	_, err = group.Array("actor", "member")
	fatalIfError(t, err)

	testGetMultStatementOfGroup(t, appendAccount(t, group))
}

func TestGetMultStatementWithActorGroupMbox(t *testing.T) {
	group, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)
	_, err = group.SetP("Group", "actor.objectType")
	fatalIfError(t, err)
	_, err = group.Array("actor", "member")
	fatalIfError(t, err)

	testGetMultStatementOfGroup(t, appendMbox(t, group))
}

func TestGetMultStatementWithActorGroupMboxSha1(t *testing.T) {
	group, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)
	_, err = group.SetP("Group", "actor.objectType")
	fatalIfError(t, err)
	_, err = group.Array("actor", "member")
	fatalIfError(t, err)

	testGetMultStatementOfGroup(t, appendMboxSHA1(t, group))
}

func TestGetMultStatementWithActorGroupOpenID(t *testing.T) {
	group, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)
	_, err = group.SetP("Group", "actor.objectType")
	fatalIfError(t, err)
	_, err = group.Array("actor", "member")
	fatalIfError(t, err)

	testGetMultStatementOfGroup(t, appendOpenID(t, group))
}

func testGetMultStatementOfGroup(t *testing.T, group *gabs.Container) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	ids := []string{
		uuid.NewV4().String(),
		uuid.NewV4().String(),
		uuid.NewV4().String(),
		uuid.NewV4().String(),
	}

	stmt, err := gabs.ParseJSON([]byte(statetmentAgentBoilerplate))
	fatalIfError(t, err)

	stmtSlice := []*gabs.Container{
		appendAccount(t, stmt),
		appendMbox(t, stmt),
		appendMboxSHA1(t, stmt),
		appendOpenID(t, stmt),
	}

	for i := 0; i < len(ids); i++ {
		putStatement(t, mart, stmtSlice[i].String(), ids[i])

		var s interface{}
		err = json.Unmarshal([]byte(stmtSlice[i].Search("actor").String()), &s)
		fatalIfError(t, err)
		err = group.ArrayAppendP(s, "actor.member")
		fatalIfError(t, err)
	}

	// construct query
	v := &url.Values{}
	//t.Log(group.Search("actor").String())
	v.Add("agent", group.Search("actor").String())

	resp := getStatement(t, mart, v)
	//t.Log(string(resp))
	respstmt, err := gabs.ParseJSON(resp)
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != len(ids) {
		t.Fatalf("Expected %d statements in response; got %d", len(ids), cnt)
	}

	children, err := respstmt.S("statements").Children()
	fatalIfError(t, err)
	for idx, stm := range children {
		if id, ok := stm.Search("id").Data().(string); !ok || id != ids[idx] {
			t.Fatalf("Got invalid order of statement array")
		}
	}
}

func postStatementWithFile(t *testing.T, mart *martini.ClassicMartini, stmt, id string) (contentSHA2sum string) {
	// construct content
	sha2 := sha256.New()
	content := bytes.NewBuffer(nil)

	// write content
	fmt.Fprintln(io.MultiWriter(content, sha2), "example content text")
	contentSHA2sum = fmt.Sprintf("%x", sha2.Sum(nil))

	// update statement
	var statement map[string]interface{}
	json.Unmarshal([]byte(stmt), &statement)
	statement["id"] = id
	statement["attachments"] = []map[string]interface{}{
		{
			"usageType": "http://example.com/attachment-usage/test",
			"display": map[string]interface{}{
				"en-US": "A test attachment",
			},
			"description": map[string]interface{}{
				"en-US": "A test attachment (description)",
			},
			"contentType": "text/plain; charset=ascii",
			"length":      content.Len(),
			"sha2":        contentSHA2sum,
		},
	}
	ustmt, _ := json.Marshal(statement)

	//
	// create multipart/form-data
	var header textproto.MIMEHeader
	buffer := bytes.NewBuffer(nil)
	encoder := multipart.NewWriter(buffer)

	// json field
	header = make(textproto.MIMEHeader)
	header.Add("Content-Type", "application/json")
	jsonfield, err := encoder.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(jsonfield, string(ustmt))

	// text (content) field
	header = make(textproto.MIMEHeader)
	header.Add("Content-Type", "text/plain")
	header.Add("Content-Transfer-Encoding", "binary")
	header.Add("X-Experience-API-Hash", contentSHA2sum)
	textfield, err := encoder.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(textfield, content)

	// finish writing
	encoder.Close()

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test/test/statements", buffer)
	req.Header.Add("Content-Type", "multipart/mixed; boundary="+encoder.Boundary())
	req.Header.Add("X-Experience-API-Version", "1.0.2")

	mart.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusOK; got != expected {
		r, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Expected %v response code from post single statement with file; got %d, %v", expected, got, string(r))
	}

	return
}

func TestGetStatementWithFile(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatement01))
	fatalIfError(t, err)

	vsuffix := uuid.NewV4().String()
	verbID := "http://example.com/realglobe/XAPIprofile/test-" + vsuffix
	_, err = stmt.SetP(verbID, "verb.id")
	fatalIfError(t, err)

	contentSHA2sum := postStatementWithFile(t, mart, stmt.String(), id1)

	// construct query
	v := &url.Values{}
	v.Add("verb", verbID)
	v.Add("attachments", "true")

	resp, head := getStatementWithHeader(t, mart, v)
	mt, ps, err := mime.ParseMediaType(head.Get("Content-Type"))
	fatalIfError(t, err)
	boundary, ok := ps["boundary"]
	if mt != "multipart/mixed" {
		t.Fatal("header is not multipart/mixed")
	}
	if !ok {
		t.Fatal("multipart/mixed, boundary not found")
	}

	// check statement part
	r := multipart.NewReader(bytes.NewReader(resp), boundary)
	p, err := r.NextPart()
	fatalIfError(t, err)

	if got, expected := p.Header.Get("Content-Type"), "application/json"; got != expected {
		t.Fatalf("Expected Content-Type: %v; got %v", expected, got)
	}

	mainJSON, err := ioutil.ReadAll(p)
	fatalIfError(t, err)
	respstmt, err := gabs.ParseJSON(mainJSON)
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 1 {
		t.Fatalf("Expected %d statements in response; got %d", 1, cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid statement")
	}

	// check content part
	p, err = r.NextPart()
	hash := p.Header.Get("X-Experience-API-Hash")
	if hash != contentSHA2sum {
		t.Fatal("Content hash does not same between request and response")
	}
}

var singleStatementWithMember = `
{
    "actor": {
        "objectType": "Group",
        "name": "Test Agent Boilerplate",
        "mbox": "mailto:test.group@realglobe.example.com",
        "member": [
            {
                "name": "Andrew Downes",
                "mbox": "mailto: andrew@realglobe.example.com",
                "objectType": "Agent"
            }
        ]
    },
    "verb": {
        "id": "http: //www.adlnet.gov/XAPIprofile/ran(travelled_a_distance)",
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

func TestGetMultStatementWithRelatedAgentsAccount(t *testing.T) {
	suffix := uuid.NewV4().String()
	name := "name-example-" + suffix
	homePage := "http://example.com"

	testGetMultStatementWithRelatedAgents(t, map[string]interface{}{
		"objectType": "Agent",
		"name":       "Aaron Silvers",
		"account": map[string]interface{}{
			"name":     name,
			"homePage": homePage,
		},
	})
}

func TestGetMultStatementWithRelatedAgentsMbox(t *testing.T) {
	domainPrefix := uuid.NewV4().String()
	mbox := "mbox:test@" + domainPrefix + ".example.com"

	testGetMultStatementWithRelatedAgents(t, map[string]interface{}{
		"objectType": "Agent",
		"name":       "Aaron Silvers",
		"mbox":       mbox,
	})
}

func TestGetMultStatementWithRelatedAgentsMboxSHA1(t *testing.T) {
	data := make([]byte, 16)
	_, err := rand.Read(data)
	fatalIfError(t, err)
	mboxSHA1 := fmt.Sprintf("%x", sha1.Sum(data))

	testGetMultStatementWithRelatedAgents(t, map[string]interface{}{
		"objectType":   "Agent",
		"name":         "Aaron Silvers",
		"mbox_sha1sum": mboxSHA1,
	})
}

func TestGetMultStatementWithRelatedAgentsOpenID(t *testing.T) {
	openIDprefix := uuid.NewV4().String()
	openid := "http://" + openIDprefix + ".openid.example.com"

	testGetMultStatementWithRelatedAgents(t, map[string]interface{}{
		"objectType": "Agent",
		"name":       "Aaron Silvers",
		"openid":     openid,
	})
}

func testGetMultStatementWithRelatedAgents(t *testing.T, agent map[string]interface{}) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	id1 := uuid.NewV4().String()
	id2 := uuid.NewV4().String()

	stmt, err := gabs.ParseJSON([]byte(singleStatementWithMember))
	fatalIfError(t, err)

	dprefix := uuid.NewV4().String()
	mailto := "mailto:test.group@" + dprefix + ".realglobe.example.com"
	_, err = stmt.SetP(mailto, "actor.mbox")
	fatalIfError(t, err)

	actor, err := gabs.Consume(agent)
	fatalIfError(t, err)
	err = stmt.ArrayAppendP(actor.Data(), "actor.member")
	fatalIfError(t, err)

	putStatement(t, mart, stmt.String(), id1)
	putStatement(t, mart, stmt.String(), id2)

	// construct query
	v := &url.Values{}
	v.Add("agent", actor.String())
	v.Add("related_agents", "true")

	respstmt, err := gabs.ParseJSON(getStatement(t, mart, v))
	fatalIfError(t, err)
	cnt, err := respstmt.ArrayCount("statements")
	fatalIfError(t, err)

	if cnt != 2 {
		t.Fatalf("Expected 2 statements in response; got %d", cnt)
	}

	s0, err := respstmt.ArrayElement(0, "statements")
	fatalIfError(t, err)
	if id, ok := s0.Search("id").Data().(string); !ok || id != id1 {
		t.Fatalf("Got invalid order of statement array")
	}

	s1, err := respstmt.ArrayElement(1, "statements")
	fatalIfError(t, err)
	if id, ok := s1.Search("id").Data().(string); !ok || id != id2 {
		t.Fatalf("Got invalid order of statement array")
	}
}

func getStatementWithUnusedStatementID(t *testing.T) {
	db := initDatabase(t)
	defer db.Close()
	mart := initHandler(db)

	unusedStatementID := "ffffffff-ffff-ffff-ffff-ffffffffffff"

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test/test/statements?statementId="+unusedStatementID, nil)
	fatalIfError(t, err)
	req.Header.Add("X-Experience-API-Version", "1.0.2")
	mart.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusNotFound; got != expected {
		t.Fatalf("Expected %v response code from get statement(s); got %d", expected, got)
	}
}
