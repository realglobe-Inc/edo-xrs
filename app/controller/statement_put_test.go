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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/acceptlang"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
)

var putStatementTestCases = []string{
	singleStatement01,
	singleStatement02,
	singleStatement03,
}

func TestPutStatement(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Put("/:user/:app/statements", c.StoreStatement)

	for _, stmt := range putStatementTestCases {
		resp := httptest.NewRecorder()
		id := uuid.NewV4().String()
		req, _ := http.NewRequest("PUT", "/test/test/statements?statementId="+id, strings.NewReader(stmt))
		req.Header.Add("X-Experience-API-Version", "1.0.2")

		m.ServeHTTP(resp, req)

		if got, expected := resp.Code, http.StatusNoContent; got != expected {
			t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
		}
	}
}

func TestPutStatementWithInvalidXAPIHeader(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Put("/:user/:app/statements", c.StoreStatement)
	stmt := singleStatement01

	resp := httptest.NewRecorder()
	id := uuid.NewV4().String()
	req, _ := http.NewRequest("PUT", "/test/test/statements?statementId="+id, strings.NewReader(stmt))
	req.Header.Add("X-Experience-API-Version", "0.0.0")

	m.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusBadRequest; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}
}

func TestPutStatementWithInvalidStatementID(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Put("/:user/:app/statements", c.StoreStatement)
	stmt := singleStatement01

	resp := httptest.NewRecorder()
	invalidID := "xxx"
	req, _ := http.NewRequest("PUT",
		"/test/test/statements?statementId="+invalidID,
		strings.NewReader(stmt),
	)
	req.Header.Add("X-Experience-API-Version", "1.0.2")

	m.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusBadRequest; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}
}

func TestPutStatementWithMismatchID(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Put("/:user/:app/statements", c.StoreStatement)
	stmt, err := gabs.ParseJSON([]byte(singleStatement01))
	if err != nil {
		t.Fatal(err)
	}

	_, err = stmt.SetP(uuid.NewV4().String(), "id")
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT",
		"/test/test/statements?statementId="+uuid.NewV4().String(),
		strings.NewReader(stmt.String()),
	)
	req.Header.Add("X-Experience-API-Version", "1.0.2")

	m.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusBadRequest; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}
}

func TestPutStatementWithConflictID(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Put("/:user/:app/statements", c.StoreStatement)
	stmt1, err := gabs.ParseJSON([]byte(singleStatement01))
	if err != nil {
		t.Fatal(err)
	}
	stmt2, err := gabs.ParseJSON([]byte(singleStatement02))
	if err != nil {
		t.Fatal(err)
	}

	// set same ID
	statementID := uuid.NewV4().String()
	_, err = stmt1.SetP(statementID, "id")
	if err != nil {
		t.Fatal(err)
	}
	_, err = stmt2.SetP(statementID, "id")
	if err != nil {
		t.Fatal(err)
	}

	// 1
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT",
		"/test/test/statements?statementId="+statementID,
		strings.NewReader(stmt1.String()),
	)
	req.Header.Add("X-Experience-API-Version", "1.0.2")

	m.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusNoContent; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}

	// 2
	resp = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT",
		"/test/test/statements?statementId="+statementID,
		strings.NewReader(stmt2.String()),
	)
	req.Header.Add("X-Experience-API-Version", "1.0.2")

	m.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusConflict; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}
}

func TestPutStatementAndCheckAuthority(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Get("/:user/:app/statements", acceptlang.Languages(), c.FindStatement)
	m.Put("/:user/:app/statements", c.StoreStatement)
	stmt := singleStatement01

	resp := httptest.NewRecorder()
	statementID := uuid.NewV4().String()
	req, _ := http.NewRequest("PUT",
		"/test/test/statements?statementId="+statementID,
		strings.NewReader(stmt),
	)
	req.Header.Add("X-Experience-API-Version", "1.0.2")
	req.Header.Add("X-Edo-Ta-Id", "example-ta-id")
	req.Header.Add("X-Edo-User-Id", "example-user-id")

	m.ServeHTTP(resp, req)

	if got, expected := resp.Code, http.StatusNoContent; got != expected {
		t.Fatalf("Expected %v response code from put single statement; got %d", expected, got)
	}

	resp2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test/test/statements?statementId="+statementID, nil)
	req2.Header.Add("X-Experience-API-Version", "1.0.2")

	m.ServeHTTP(resp2, req2)

	if got, expected := resp2.Code, http.StatusOK; got != expected {
		t.Fatalf("Expected %v response code from get single statement; got %d", expected, got)
	}

	body, err := ioutil.ReadAll(resp2.Body)
	fatalIfError(t, err)

	rstmt, err := gabs.ParseJSON(body)
	fatalIfError(t, err)

	v, ok := rstmt.Path("authority.account.homePage").Data().(string)
	if !ok || v != "example-ta-id" {
		t.Fatal("Field authority.account.homePage is invalid or not found in get response")
	}
	v, ok = rstmt.Path("authority.account.name").Data().(string)
	if !ok || v != "example-user-id" {
		t.Fatal("Field authority.account.name is invalid or not found in get response")
	}
}
