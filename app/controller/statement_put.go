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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/go-martini/martini"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"github.com/realglobe-Inc/edo-xrs/app/model"
	"github.com/realglobe-Inc/edo-xrs/app/validator"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// StoreStatement はステートメントの PUT リクエストを扱うハンドラである。
// このハンドラは与えられたステートメントをバリデートし、データベースに挿入する。
// URLパラメータには UUID (ステートメントID) が与えられており、そのIDのステートメントを挿入する。
func (c *Controller) StoreStatement(params martini.Params, w http.ResponseWriter, req *http.Request) (int, string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Experience-API-Version", "1.0.2")
	user, app := params["user"], params["app"]

	contentType := req.Header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = "application/json"
	}
	statements, attachmentSHA2s, err := c.parseRequestBody(req.Body, contentType)
	if err != nil {
		return NewBadRequestErrF("An error occured on parse request: %s", err).Response()
	}
	if len(statements) != 1 {
		return NewBadRequestErr("Statement must be single on PUT request").Response()
	}

	// parseRequestBody でチェックしているため変換可能
	statement := statements[0].(map[string]interface{})

	// xAPI のバージョンを確認, Experience API, Section 6.2 を参照
	xAPIVersion := req.Header.Get("X-Experience-API-Version")
	if err := validator.Statement(validator.ToXAPIVersion(xAPIVersion), statement); err != nil {
		return NewBadRequestErrF("Invalid statement: %s", err).Response()
	}

	// ステートメントIDをチェック, Experience API, Section 7.2.1 を参照
	statementID := req.URL.Query().Get("statementId")
	if !validator.IsUUID(statementID) {
		return NewBadRequestErr("Statement ID must be valid UUID").Response()
	}

	// Attachment と multipart/mixed に指定されるヘッダ情報との整合性をチェック
	if attachmentSHA2s != nil && !hasSameHashBetween(statements, attachmentSHA2s) {
		return NewBadRequestErr("Unexpected content hash given on attachment or multipart header").Response()
	}

	// 与えられたステートメントと、URLのパラメータのIDをチェック
	if id, ok := statement["id"]; !ok {
		// URL にのみ ID が付加されているときは, ステートメントにその ID を補完
		statement["id"] = statementID
	} else if id.(string) != statementID {
		// ステートメントに指定されている ID と違っている場合はエラー
		return NewBadRequestErr("ID mismatch between URL parameter and request body").Response()
	}

	// タイムスタンプに関する処理
	currentTime := time.Now()

	timestamp := currentTime
	statement["stored"] = currentTime

	if ts, ok := statement["timestamp"]; ok {
		t, err := time.Parse(time.RFC3339Nano, ts.(string))
		if err != nil {
			return NewBadRequestErr("Timestamp must be of the form of RFC3339").Response()
		}
		timestamp = t
	}

	if code, mess := c.insertIntoDB(xAPIVersion, user, app, model.DocumentSlice{
		*model.NewDocument(xAPIVersion, user, app, timestamp, statement),
	}); code != http.StatusOK {
		return code, mess
	}

	return http.StatusNoContent, "No Content"
}

func hasSameHashBetween(statements []interface{}, attachmentSHA2s []string) bool {
	sha2map := make(map[string]bool)
	for _, stmt := range statements {
		if attachments, ok := (stmt.(map[string]interface{}))["attachments"]; ok {
			for _, att := range attachments.([]interface{}) {
				a := att.(map[string]interface{})
				if sha2, ok := a["sha2"]; ok {
					sha2map[sha2.(string)] = true
				}
			}
		}
	}
	for _, sha2 := range attachmentSHA2s {
		if !sha2map[sha2] {
			return false
		}
	}

	return len(sha2map) == len(attachmentSHA2s)
}

// StoreMultStatement はステートメントを単一、もしくは複数挿入するためのハンドラである。
// ステートメントは配列、もしくは単一のJSONの形でリクエストボディに与えられる。
func (c *Controller) StoreMultStatement(params martini.Params, w http.ResponseWriter, req *http.Request) (int, string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Experience-API-Version", "1.0.2")
	user, app := params["user"], params["app"]

	contentType := req.Header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = "application/json"
	}
	statements, attachmentSHA2s, err := c.parseRequestBody(req.Body, contentType)
	if err != nil {
		return NewBadRequestErrF("An error occured on parse request: %s", err).Response()
	}

	// xAPI のバージョンを確認, Experience API, Section 6.2 を参照
	xAPIVersion := req.Header.Get("X-Experience-API-Version")
	if err := validator.MultStatement(validator.ToXAPIVersion(xAPIVersion), statements); err != nil {
		return NewBadRequestErrF("Invalid statement: %s", err).Response()
	}

	// attachments がある場合、チェック
	if attachmentSHA2s != nil && !hasSameHashBetween(statements, attachmentSHA2s) {
		return NewBadRequestErr("Unexpected content hash given on attachment or multipart header").Response()
	}

	docs, insertedIDs, err := parseStatementsAndGetIDs(xAPIVersion, user, app, statements)
	if err != nil {
		return NewBadRequestErrF("Invalid statements: %s", err).Response()
	}

	if status, mess := c.insertIntoDB(xAPIVersion, user, app, docs); status != http.StatusOK {
		return status, mess
	}

	result, err := json.Marshal(insertedIDs)
	if err != nil {
		logger.Err("An unexpected error occured: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	w.Header().Set("Content-Type", "application/json")
	return http.StatusOK, string(result)
}

func (c *Controller) insertIntoDB(xAPIVersion, user, app string, docs model.DocumentSlice) (int, string) {
	// データベースのセッションを取得
	sess := c.session.New()
	defer sess.Close()
	db := sess.DB(miscs.GlobalConfig.MongoDB.DBName)

	if valid, err := isValidVoidedStatements(db, docs, xAPIVersion, user, app); !valid && err != nil {
		return NewBadRequestErrF("Invalid voided statement: %s", err).Response()
	}

	quota, err := model.GetQuota(db, user)
	if err != nil {
		logger.Err("An unexpected error occured on get quota: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	// ユーザーのディスク使用量をチェック
	if !quota.Check() {
		return NewBadRequestErrF("The disk is full of user: %s", user).Response()
	}

	if err := quota.IncrementUsageTo(db, getSizeOfDocuments(docs)); err != nil {
		logger.Err("An unexpected error occured on increment quota: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	// データベースに挿入し、duplicate key エラーが発生しない場合に正常終了、
	// そうでなければ Conflict を返す。xAPI の仕様によると Conflict は
	// statement の id フィールド値が重複する場合と規定されている。
	// そこで、この箇所では statement の id フィールドを unique index にすることで
	// duplicate key エラーを発生させ、Conflict の判定を行っている。
	if err := docs.InsertTo(db.C("statement")); err != nil {
		if !mgo.IsDup(err) {
			logger.Err("An unexpected error occured on insert statement into DB: ", err)
			return http.StatusInternalServerError, "Internal Server Error"
		}

		return http.StatusConflict, "Conflict"
	}

	return http.StatusOK, "ok"
}

func getSizeOfDocuments(docs model.DocumentSlice) int64 {
	var total int64

	for _, doc := range docs {
		if res, err := json.Marshal(doc.Data); err == nil {
			total += int64(len(res))
		} else {
			logger.Err("An unexpected error occured:", err)
		}
	}

	return total
}

func isValidVoidedStatement(db *mgo.Database, doc *model.Document, xAPIVersion, user, app string) (bool, error) {
	reqBody := doc.Data

	// ステートメントが Voided である時の処理
	if verb, ok := reqBody["verb"].(map[string]interface{}); ok {
		if verb["id"].(string) == miscs.GlobalConfig.Global.VoidedStatementID {
			object, ok := reqBody["object"].(map[string]interface{})
			if !ok || object["objectType"] != "StatementRef" {
				return false, errors.New("verb.objectType must be StatementRef on voided statement")
			}

			// Voided に Voided を被せる時はエラーを返す
			if isVoidedStatement(db.C("statement"), xAPIVersion, user, app, object["id"].(string)) {
				return false, errors.New("voided statement cannot be voided")
			}
		}
	}

	return true, nil
}

func isValidVoidedStatements(db *mgo.Database, docs model.DocumentSlice, xAPIVersion, user, app string) (bool, error) {
	for _, doc := range docs {
		if valid, err := isValidVoidedStatement(db, &doc, xAPIVersion, user, app); !valid {
			return valid, err
		}
	}

	return true, nil
}

func (c *Controller) parseRequestBody(r io.Reader, t string) ([]interface{}, []string, error) {
	mediatype, params, err := mime.ParseMediaType(t)
	if err != nil {
		return nil, nil, err
	}

	var statements []interface{}

	// リクエストボディが multipart/mixed の場合, ボディの中身はデータベースに入れて
	// その sha2 値を collect する。また, json 値がきたときにはステートメントとする
	if mediatype == "multipart/mixed" {
		var sha2slice []string
		boundary, ok := params["boundary"]
		if !ok {
			return nil, nil, errors.New("invalid or no multipart boundary in Content-Type")
		}

		sha2 := sha256.New()
		sess := c.session.New()
		defer sess.Close()
		gfs := sess.DB(miscs.GlobalConfig.MongoDB.DBName).GridFS("fs")
		mr := multipart.NewReader(r, boundary)

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, nil, err
			}
			mt, _, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
			switch mt {
			case "application/json":
				stmts, err := readBodyJSONOfArray(p)
				if err != nil {
					return nil, nil, err
				}
				statements = append(statements, stmts...)
			default:
				hash := p.Header.Get("X-Experience-API-Hash")
				if len(hash) == 0 {
					return nil, nil, errors.New("X-Experience-API-Hash is empty or not specified")
				}
				gfsfile, err := gfs.Create(hash)
				if err != nil {
					return nil, nil, err
				}
				gfsfile.SetMeta(bson.M{
					"Content-Type":              p.Header.Get("Content-Type"),
					"Content-Transfer-Encoding": p.Header.Get("Content-Transfer-Encoding"),
				})
				_, err = io.Copy(io.MultiWriter(gfsfile, sha2), p)
				if err != nil { // NOTE: A successful Copy returns err == nil, not err == os.EOF
					return nil, nil, err
				}
				if err = gfsfile.Close(); err != nil {
					return nil, nil, err
				}
				if fmt.Sprintf("%x", sha2.Sum(nil)) != hash {
					gfs.RemoveId(gfsfile.Id())
					return nil, nil, errors.New("content hash and X-Experiece-API-Hash does not match")
				}
				sha2slice = append(sha2slice, hash)

				sha2.Reset()
			}
		}
		if len(statements) == 0 {
			return nil, nil, errors.New("statement is empty on multipart content")
		}
		return statements, sha2slice, nil
	}

	// when mediatype == "application/json", then
	statements, err = readBodyJSONOfArray(r)
	if err != nil {
		return nil, nil, err
	}
	return statements, nil, err
}

// readBodyJSON は与えられたリクエストボディを読み、それを JSON としてパースした結果を返す。
// リクエストボディが読めなかったり、与えられた値が JSON でない場合は err != nil となる。
func readBodyJSON(r io.Reader) (interface{}, error) {
	var result interface{}
	err := json.NewDecoder(r).Decode(&result)

	if err != nil {
		return nil, errors.New("invalid request body")
	}

	return result, nil
}

func readBodyJSONOfArray(r io.Reader) ([]interface{}, error) {
	// リクエストボディをパース
	var reqBody []interface{}
	if b, err := readBodyJSON(r); err == nil {
		// リクエストにはステートメントの配列、もしくは単一のステートメントが入る。
		// そのため、後述のコードの単純化のため、単一である場合、一つの要素を持つ配列として扱う。
		switch t := b.(type) {
		case map[string]interface{}:
			reqBody = []interface{}{t}
		case []interface{}:
			reqBody = t
		default:
			return nil, fmt.Errorf("invalid JSON structure, unexpected type %T", t)
		}
	} else {
		return nil, err
	}

	return reqBody, nil
}

func parseStatementsAndGetIDs(version, user, app string, reqBody []interface{}) (model.DocumentSlice, []string, error) {
	docs := make(model.DocumentSlice, 0, len(reqBody))
	insertedIDs := make([]string, 0, len(reqBody))
	currentTime := time.Now()

	for _, v := range reqBody {
		stmt, ok := v.(map[string]interface{})
		if !ok {
			return nil, nil, errors.New("unexpected statement structure")
		}

		// ID 入っていなければ補完
		if _, ok := stmt["id"]; !ok {
			stmt["id"] = uuid.NewV4().String()
		}

		// タイムスタンプに関する処理
		timestamp := currentTime
		stmt["stored"] = currentTime

		if ts, ok := stmt["timestamp"]; ok {
			t, err := time.Parse(time.RFC3339Nano, ts.(string))

			if err != nil {
				return nil, nil, fmt.Errorf("timestamp must be of the form of RFC3339")
			}
			timestamp = t
		}

		docs = append(docs, *model.NewDocument(version, user, app, timestamp, stmt))
		insertedIDs = append(insertedIDs, stmt["id"].(string))
	}

	return docs, insertedIDs, nil
}
