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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/go-martini/martini"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"gopkg.in/mgo.v2"
)

var postStatementWithFileTestCases = []string{
	singleStatement01,
}

func TestPostStatementWithFile(t *testing.T) {
	m := martini.Classic()

	sess, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer sess.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	c := New(sess)

	m.Post("/:user/:app/statements", c.StoreMultStatement)

	for _, stmt := range postStatementWithFileTestCases {
		// construct content
		sha2 := sha256.New()
		content := bytes.NewBuffer(nil)

		// write content
		fmt.Fprintln(io.MultiWriter(content, sha2), "example content text")
		contentSha2sum := fmt.Sprintf("%x", sha2.Sum(nil))

		// update statement
		var statement map[string]interface{}
		json.Unmarshal([]byte(stmt), &statement)
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
				"sha2":        contentSha2sum,
			},
		}
		ustmt, _ := json.Marshal(statement)

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
		header.Add("X-Experience-API-Hash", contentSha2sum)
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

		m.ServeHTTP(resp, req)

		if got, expected := resp.Code, http.StatusOK; got != expected {
			r, _ := ioutil.ReadAll(resp.Body)
			t.Fatalf("Expected %v response code from post single statement with file; got %d, %v", expected, got, string(r))
		}
	}
}
