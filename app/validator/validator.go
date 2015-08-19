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

package validator

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/realglobe-Inc/go-lib/rglog"
	"github.com/realglobe-Inc/gojsonschema"
)

type XAPISchema map[XAPIVersion]map[string]*gojsonschema.Schema

var (
	logger = rglog.Logger("XAPI/validator/validator")
	schema = &XAPISchema{}
)

func (s *XAPISchema) Get(version XAPIVersion, name string) *gojsonschema.Schema {
	return (*s)[version][name]
}

func GetXAPISchemaInstance() *XAPISchema {
	return schema
}

func init() {
	var err error

	var schemaPathV10x string
	gopath := os.Getenv("GOPATH")
	if len(gopath) != 0 {
		schemaPathV10x = gopath + "/src/github.com/realglobe-Inc/edo-xrs/jsonschema/xapi_1.0.2"
	} else {
		if schemaPathV10x, err = filepath.Abs("jsonschema/xapi_1.0.2"); err != nil {
			logger.Err("Invalid schema-path given: ", err)
			os.Exit(1)
		}
	}

	// JSON Schema をロード
	v10xstatement, err := readSchema(schemaPathV10x + "/statement.json")
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}
	v10xagent, err := readSchema(schemaPathV10x + "/agent.json")
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	v10xlangmap, err := readSchema(schemaPathV10x + "/langmap.json")
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	(*schema)[XAPIVersion10x] = map[string]*gojsonschema.Schema{
		"statement": v10xstatement,
		"agent":     v10xagent,
		"langmap":   v10xlangmap,
	}
}

// readSchema は与えられたパスの JSON Schema を読み込み、そのインスタンスを返す。
func readSchema(path string) (*gojsonschema.Schema, error) {
	var err error

	bodyText, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("Schema file not found: " + err.Error())
	}

	scm, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(bodyText)))

	if err != nil {
		return nil, errors.New("Invalid json-schema given: " + err.Error())
	}

	return scm, nil
}

// validate は与えられた body が kind であることを検査する。
func validate(version XAPIVersion, kind string, body map[string]interface{}) error {
	if s, ok := (*schema)[version]; ok {
		res, err := s[kind].Validate(gojsonschema.NewGoLoader(body))
		if err != nil {
			logger.Warn("Validation error: ", err)
			return err
		}

		if !res.Valid() {
			emessage := "The document is not valid. see errors: \n"

			for _, desc := range res.Errors() {
				emessage += fmt.Sprintf("- %s", desc)
			}
			return errors.New(emessage)
		}
	} else {
		return errors.New("Invalid XAPI version given.")
	}

	return nil
}

// Agent は body により与えられたステートメントの JSON の構造が
// xapiVersion により与えられる XAPIバージョンの正しいエージェントであることを検査する。
// ステートメントが正しくない場合、err != nil となる。
func Agent(version XAPIVersion, body map[string]interface{}) error {
	return validate(version, "agent", body)
}

// Statement は body により与えられたステートメントの JSON の構造が
// xapiVersion により与えられる XAPIバージョンの正しいステートメントであることを検査する。
// ステートメントが正しくない場合、err != nil となる。
func Statement(version XAPIVersion, body map[string]interface{}) error {
	return validate(version, "statement", body)
}

// MultStatement は body により与えられたステートメントの列を検査する。
// 配列の各ステートメントは Statement と同様に検査する。
func MultStatement(version XAPIVersion, body []interface{}) error {

	for _, stmt := range body {
		// ステートメントは必ず map[string]inteface{} の構造をしているので、それ以外はエラーを返す。
		if _, ok := stmt.(map[string]interface{}); !ok {
			return errors.New("The structure of statement doesn't satisfy specification.")
		}

		// Statement を検査。
		// MultStatement では XAPIバージョンのチェックを行っていないが、バージョンが合致しない場合、
		// Statement によりエラーが返されるのため差し障りがない。
		if err := Statement(version, stmt.(map[string]interface{})); err != nil {
			return err
		}
	}

	return nil
}
