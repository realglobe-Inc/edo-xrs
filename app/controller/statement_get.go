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
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"time"

	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/acceptlang"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"github.com/realglobe-Inc/edo-xrs/app/model"
	"github.com/realglobe-Inc/edo-xrs/app/validator"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// FindStatement handles request of search statement
func (c *Controller) FindStatement(params martini.Params,
	languages acceptlang.AcceptLanguages, w http.ResponseWriter, req *http.Request) (int, string) {
	user, app := params["user"], params["app"]

	// check version of xAPI
	xAPIVersion := req.Header.Get("X-Experience-API-Version")
	if !validator.IsValidXAPIVersion(xAPIVersion) {
		return NewBadRequestErr("Invalid or empty xAPI version given in X-Experience-API-Version").Response()
	}

	urlParams := req.URL.Query()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Experience-API-Version", "1.0.2")

	// find single statement if statementId or voidedStatementId is specified
	if len(urlParams.Get("statementId")) > 0 || len(urlParams.Get("voidedStatementId")) > 0 {
		return c.findSingleStatement(xAPIVersion, user, app, urlParams, w)
	}

	// otherwise find multiple statments
	return c.findMultipleStatements(xAPIVersion, user, app, languages, urlParams, w)
}

// findSingleStatement は単一のステートメントをレスポンスとして返す。
// この関数はリクエストパラメータに statementId もしくは voidedStatementId が指定されている時のみ呼ばれる。
func (c *Controller) findSingleStatement(xAPIVersion, user, app string, params url.Values, rw http.ResponseWriter) (int, string) {

	// リクエストパラメータのバリデート
	if !validSingleStmtRequestOf(params) {
		return NewBadRequestErr("Got invalid extra parameter with statementId or voidedStatementId").Response()
	}

	includeAttachments, err := parseParamBool(params, "attachments")
	if err != nil {
		return NewBadRequestErrF("Invalid attachments parameter given: %s", err).Response()
	}

	query := bson.M{
		"version": xAPIVersion,
		"user":    user,
		"app":     app,
	}

	// データベースのセッションを取得
	sess := c.session.New()
	defer sess.Close()
	col := sess.DB(miscs.GlobalConfig.MongoDB.DBName).C("statement")

	// statementId, もしくは voidedStatementId をチェックし、データベースのクエリを生成する。
	if statementID := params.Get("statementId"); validator.IsUUID(statementID) {
		// リクエストパラメータに statementId が指定されていたとき
		if isVoidedStatement(col, xAPIVersion, user, app, statementID) {
			// 指定されたステートメントIDが Voided ならば not found
			return http.StatusNotFound, "Not Found"
		}
		query["data.id"] = statementID
	} else if voidedStatementID := params.Get("voidedStatementId"); validator.IsUUID(voidedStatementID) {
		// リクエストパラメータに voidedStatementId が指定されていたとき
		if !isVoidedStatement(col, xAPIVersion, user, app, voidedStatementID) {
			// 指定されたステートメントIDが Voided """でなければ"" not found
			return http.StatusNotFound, "Not Found"
		}
		query["data.id"] = voidedStatementID
	} else {
		return NewBadRequestErr("A statementId or voidedStatementId is required").Response()
	}

	var document model.Document
	err = col.Find(query).One(&document)
	if err != nil {
		logger.Err("An unexpected error occured on find DB: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	res, err := json.Marshal(document.Data)
	if err != nil {
		// data はデータベースから受け取った値なので、このエラーが起きる場合は深刻な問題がある。
		logger.Err("An unexpected error occured: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	// attachment を含まないのであればそのまま返す
	if !includeAttachments {
		rw.Header().Set("Content-Type", "application/json")
		return http.StatusOK, string(res)
	}

	// attachment を含むリクエスト
	sha2s := collectSHA2sOfAttachments(model.DocumentSlice{document})
	buf, boundary, err := appendAttachments(sess, res, sha2s)
	if err != nil {
		logger.Warn(err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	rw.Header().Set("Content-Type", "multipart/mixed; boundary="+boundary)
	return http.StatusOK, string(buf)
}

func (c *Controller) findMultipleStatements(xAPIVersion, user, app string,
	languages acceptlang.AcceptLanguages, params url.Values, rw http.ResponseWriter) (int, string) {
	var queryTerms []interface{}

	if agent := params.Get("agent"); len(agent) > 0 {
		relatedAgents, err := parseParamBool(params, "related_agents")
		if err != nil {
			return NewBadRequestErrF("Invalid related_agents paramter given: %s", err).Response()
		}
		qs, err := queryOfAgent(xAPIVersion, agent, relatedAgents)
		if err != nil {
			return NewBadRequestErrF("Invalid agent given: %s", err).Response()
		}

		queryTerms = append(queryTerms, qs)
	}

	if v, ok := params["verb"]; ok && len(v) > 0 {
		queryTerms = append(queryTerms, bson.M{"data.verb.id": v[0]})
	}

	if v, ok := params["activity"]; ok && len(v) > 0 {
		id := v[0]
		activityQuery := []bson.M{}

		activityQuery = append(activityQuery, bson.M{
			"data.object.objectType": "Activity",
			"data.object.id":         id,
		})

		relatedActivities, err := parseParamBool(params, "related_activities")
		if err != nil {
			return NewBadRequestErrF("Invalid related_activities paramter given: %s", err).Response()
		}

		if relatedActivities {
			activityQuery = append(activityQuery, bson.M{
				"data.context.contextActivities.parent": bson.M{"$elemMatch": bson.M{"id": id}},
			}, bson.M{
				"data.context.contextActivities.grouping": bson.M{"$elemMatch": bson.M{"id": id}},
			}, bson.M{
				"data.context.contextActivities.category": bson.M{"$elemMatch": bson.M{"id": id}},
			}, bson.M{
				"data.context.contextActivities.other": bson.M{"$elemMatch": bson.M{"id": id}},
			})
		}
		queryTerms = append(queryTerms, bson.M{"$or": activityQuery})
	}

	if v, ok := params["registration"]; ok && len(v) > 0 {
		queryTerms = append(queryTerms, bson.M{
			"data.context.registration": v[0],
		})
	}

	if v, ok := params["since"]; ok && len(v) > 0 {
		if t, err := time.Parse(time.RFC3339Nano, v[0]); err == nil {
			queryTerms = append(queryTerms, bson.M{
				"timestamp": bson.M{"$gt": t},
			})
		} else {
			return NewBadRequestErr("Since must be of the form of RFC3339 Date/Time").Response()
		}
	}
	if v, ok := params["until"]; ok && len(v) > 0 {
		if t, err := time.Parse(time.RFC3339Nano, v[0]); err == nil {
			queryTerms = append(queryTerms, bson.M{
				"timestamp": bson.M{"$lt": t},
			})
		} else {
			return NewBadRequestErr("Until must be of the form of RFC3339 Date/Time").Response()
		}
	}

	limit := miscs.GlobalConfig.Global.MaxStatements
	if v, ok := params["limit"]; ok && len(v) > 0 {
		var err error

		if limit, err = strconv.Atoi(v[0]); err != nil {
			return NewBadRequestErr("Limit must be integer value").Response()
		}
	}

	formatType := "exact"
	if v, ok := params["format"]; ok && len(v) > 0 {
		formatType = v[0]
	}

	includeAttachments, err := parseParamBool(params, "attachments")
	if err != nil {
		return NewBadRequestErrF("Invalid attachments parameter given: %s", err).Response()
	}

	// ソート順
	sortField := "timestamp"
	ascending, err := parseParamBool(params, "ascending")
	if err != nil {
		return NewBadRequestErrF("Invalid ascending parameter given: %s", err).Response()
	}
	if ascending {
		sortField = "-timestamp"
	}

	queryTerms = append(queryTerms, bson.M{
		"version": xAPIVersion,
		"user":    user,
		"app":     app,
	})

	query := bson.M{"$and": queryTerms}

	// データベースのセッションを取得
	session := c.session.New()
	defer session.Close()
	col := session.DB(miscs.GlobalConfig.MongoDB.DBName).C("statement")

	// fetch statemsnts from DB and construct response body
	var respStatements model.DocumentSlice

	var result model.Document
	iter := col.Find(query).Sort(sortField).Limit(limit).Iter()
	for iter.Next(&result) {
		respStatements = append(respStatements, result)
	}
	if err := iter.Close(); err != nil {
		logger.Err("An unexpected error occured: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	respBody, err := formatRespMultipleStatements(formatType, languages, respStatements)
	if err != nil {
		logger.Err("An unexpected error occured: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	// attachment を含まないのであればそのまま返す
	if !includeAttachments {
		rw.Header().Set("Content-Type", "application/json")
		return http.StatusOK, string(respBody)
	}

	// attachment を含むリクエスト
	sha2s := collectSHA2sOfAttachments(respStatements)
	buf, boundary, err := appendAttachments(session, respBody, sha2s)
	if err != nil {
		logger.Err("An unexpected error occured: ", err)
		return http.StatusInternalServerError, "Internal Server Error"
	}

	rw.Header().Set("Content-Type", "multipart/mixed; boundary="+boundary)
	return http.StatusOK, string(buf)
}

func isVoidedStatement(collection *mgo.Collection, version, user, app, id string) bool {
	query := bson.M{
		"version":                version,
		"user":                   user,
		"app":                    app,
		"data.verb.id":           miscs.GlobalConfig.Global.VoidedStatementID,
		"data.object.objectType": "StatementRef",
		"data.object.id":         id,
	}

	count, err := collection.Find(query).Count()
	if err != nil {
		logger.Err("An unexpected error occured on find voided statement in DB: ", err)
		return false
	}

	return count > 0
}

// validSingleStmtRequestOf は引数にリクエストパラメータを与えることでその
// パラメータが仕様の形を満たしていることをチェックする。(Experience API, Section 7.2.3)
// 具体的には, statementId, もしくは voidedStatementId のどちらかと
// attachments, format が与えられたものが正しいパラメータであり, それ以外を含む場合はエラーとなる
func validSingleStmtRequestOf(values url.Values) bool {
	var result = false

	if v, ok := values["statementId"]; ok {
		result = len(v) > 0
	}
	if v, ok := values["voidedStatementId"]; ok {
		result = !result && len(v) > 0
	}
	if !result {
		return false
	}

	var count = 1
	if _, ok := values["attachments"]; ok {
		count++
	}
	if _, ok := values["format"]; ok {
		count++
	}

	return len(values) <= count
}

func queryOfTerms(terms []map[string]interface{}) bson.M {
	var identTerms []bson.M

	for _, term := range terms {
		for k, v := range term {
			identTerms = append(identTerms, bson.M{k: v})
		}
	}

	return bson.M{"$or": identTerms}
}

func constructIFITerms(prefix string, agent map[string]interface{}) (terms []map[string]interface{}) {
	if account, ok := agent["account"]; ok {
		terms = append(terms, map[string]interface{}{
			"$and": []bson.M{
				{prefix + ".account.name": (account.(map[string]interface{}))["name"].(string)},
				{prefix + ".account.homePage": (account.(map[string]interface{}))["homePage"].(string)},
			},
		})
	}
	if mbox, ok := agent["mbox"]; ok {
		terms = append(terms, map[string]interface{}{prefix + ".mbox": mbox.(string)})
	}
	if mboxSHA1, ok := agent["mbox_sha1sum"]; ok {
		terms = append(terms, map[string]interface{}{prefix + ".mbox_sha1sum": mboxSHA1.(string)})
	}
	if openID, ok := agent["openid"]; ok {
		terms = append(terms, map[string]interface{}{prefix + ".openid": openID.(string)})
	}

	return
}

func constructIFITermsOfArray(prefix string, agent map[string]interface{}) (terms []map[string]interface{}) {
	if account, ok := agent["account"]; ok {
		terms = append(terms, map[string]interface{}{
			"$and": []bson.M{
				{
					prefix: bson.M{
						"$elemMatch": bson.M{
							"account.name": (account.(map[string]interface{}))["name"].(string),
						},
					},
				},
				{
					prefix: bson.M{
						"$elemMatch": bson.M{
							"account.homePage": (account.(map[string]interface{}))["homePage"].(string),
						},
					},
				},
			},
		})
	}
	if mbox, ok := agent["mbox"]; ok {
		terms = append(terms, map[string]interface{}{
			prefix: bson.M{"$elemMatch": bson.M{"mbox": mbox.(string)}},
		})
	}
	if mboxSHA1, ok := agent["mbox_sha1sum"]; ok {
		terms = append(terms, map[string]interface{}{
			prefix: bson.M{"$elemMatch": bson.M{"mbox_sha1sum": mboxSHA1.(string)}},
		})
	}
	if openID, ok := agent["openid"]; ok {
		terms = append(terms, map[string]interface{}{
			prefix: bson.M{"$elemMatch": bson.M{"openid": openID.(string)}},
		})
	}

	return
}

func ifiTermsOfAgent(agent map[string]interface{}, isRelated bool) (terms []map[string]interface{}) {
	terms = append(terms, constructIFITerms("data.actor", agent)...)

	if isRelated {
		// statement
		terms = append(terms, constructIFITermsOfArray("data.actor.member", agent)...)
		terms = append(terms, constructIFITerms("data.object", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.object.member", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.authority.member", agent)...)
		terms = append(terms, constructIFITerms("data.context.instructor", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.context.instructor.member", agent)...)
		terms = append(terms, constructIFITerms("data.context.team", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.context.team.member", agent)...)
		// substatement
		terms = append(terms, constructIFITerms("data.object.actor", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.object.actor.member", agent)...)
		terms = append(terms, constructIFITerms("data.object.object", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.object.object.member", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.object.authority.member", agent)...)
		terms = append(terms, constructIFITerms("data.object.context.instructor", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.object.context.instructor.member", agent)...)
		terms = append(terms, constructIFITerms("data.object.context.team", agent)...)
		terms = append(terms, constructIFITermsOfArray("data.object.context.team.member", agent)...)
	}

	return
}

func termsOfAgent(agent map[string]interface{}, isRelated bool) (terms []map[string]interface{}) {
	terms = append(terms, ifiTermsOfAgent(agent, isRelated)...)

	if memberSlice, ok := agent["member"]; ok {
		for _, m := range memberSlice.([]interface{}) {
			member := m.(map[string]interface{})

			terms = append(terms, ifiTermsOfAgent(member, isRelated)...)
		}
	}

	return
}

// queryOfAgent は agnet の文字列を受け取り、データベースのクエリを返す。
// 引数に与えられた文字列が変換不可ならエラーを返す。
func queryOfAgent(xAPIVersion, agentString string, relatedActivities bool) (bson.M, error) {
	var agent map[string]interface{}
	err := json.Unmarshal([]byte(agentString), &agent)

	if err != nil {
		return nil, err
	}

	if err = validator.Agent(validator.ToXAPIVersion(xAPIVersion), agent); err != nil {
		return nil, err
	}

	terms := termsOfAgent(agent, relatedActivities)

	return queryOfTerms(terms), nil
}

// URLパラーメタの field に指定されたキーに格納されている bool 値を取得する。
// 指定されたキーに値がなかった場合にはデフォルト値 false が返される。
func parseParamBool(params url.Values, field string) (bool, error) {
	var result = false

	if v, ok := params[field]; ok && len(v) > 0 {
		var err error

		if result, err = strconv.ParseBool(v[0]); err != nil {
			return false, fmt.Errorf("parse error: %s", err)
		}
	}

	return result, nil
}

func projectMapSingle(object map[string]interface{}, keys []string) map[string]interface{} {
	ran := make(map[string]interface{})
	for _, key := range keys {
		if content, ok := object[key]; ok {
			ran[key] = content
			break
		}
	}
	return ran
}

func projectDescriptionInArray(root []interface{}, keys []string) []map[string]interface{} {
	var cs []map[string]interface{}

	for _, c := range root {
		m := c.(map[string]interface{})
		desc, ok := m["description"]
		if ok {
			m["description"] = projectMapSingle(
				desc.(map[string]interface{}),
				keys,
			)
		}

		cs = append(cs, m)
	}

	return cs
}

// canonical フォーマットが指定されたとき、優先して取得する言語のリスト.
// リストの先頭が最も優先順位が高い
var canonFormatFilterPriority = []string{"ja-JP", "en-US"}

func formatCanonicalResponse(docs model.DocumentSlice, languages acceptlang.AcceptLanguages) ([]byte, error) {
	var formatFilterPriori []string
	if languages.Len() > 0 {
		var langstr []string
		for _, lang := range languages {
			langstr = append(langstr, lang.Language)
		}
		formatFilterPriori = append(langstr, canonFormatFilterPriority...)
	} else {
		formatFilterPriori = canonFormatFilterPriority
	}

	// Language Map 型をもつフィールドの言語の種類を一つに絞る
	docs.Map(func(d model.Document) model.Document {
		// TODO: 工夫すれば、Mongo の Projection 機能でいけるかもしれない
		if verb, ok := d.Data["verb"].(map[string]interface{}); ok {
			if display, ok := verb["display"].(map[string]interface{}); ok {
				verb["display"] = projectMapSingle(display, formatFilterPriori)
			}
		}
		if object, ok := d.Data["object"].(map[string]interface{}); ok {
			if objectDefinition, ok := object["definition"].(map[string]interface{}); ok {
				if name, ok := objectDefinition["name"].(map[string]interface{}); ok {
					objectDefinition["name"] =
						projectMapSingle(name, formatFilterPriori)
				}
				if description, ok := objectDefinition["description"].(map[string]interface{}); ok {
					objectDefinition["description"] =
						projectMapSingle(description, formatFilterPriori)
				}
				if choices, ok := objectDefinition["choices"].([]interface{}); ok {
					objectDefinition["choices"] =
						projectDescriptionInArray(choices, formatFilterPriori)
				}
				if scale, ok := objectDefinition["scale"].([]interface{}); ok {
					objectDefinition["scale"] =
						projectDescriptionInArray(scale, formatFilterPriori)
				}
				if source, ok := objectDefinition["source"].([]interface{}); ok {
					objectDefinition["source"] =
						projectDescriptionInArray(source, formatFilterPriori)
				}
				if target, ok := objectDefinition["target"].([]interface{}); ok {
					objectDefinition["target"] =
						projectDescriptionInArray(target, formatFilterPriori)
				}
				if steps, ok := objectDefinition["steps"].([]interface{}); ok {
					objectDefinition["steps"] =
						projectDescriptionInArray(steps, formatFilterPriori)
				}
			}
		}
		if attachments, ok := d.Data["attachments"].(map[string]interface{}); ok {
			if display, ok := attachments["display"].(map[string]interface{}); ok {
				attachments["display"] =
					projectMapSingle(display, formatFilterPriori)
			}
			if display, ok := attachments["description"].(map[string]interface{}); ok {
				attachments["description"] =
					projectMapSingle(display, formatFilterPriori)
			}
		}

		return d
	})

	return json.Marshal(map[string]interface{}{
		"statements": docs.ToDataArray(),
		"more":       "",
	})
}

func formatExactResponse(docs model.DocumentSlice) ([]byte, error) {
	var respData []interface{}
	for _, doc := range docs {
		respData = append(respData, doc.Data)
	}
	return json.Marshal(map[string]interface{}{
		"statements": respData,
		"more":       "",
	})
}

func formatIDsResponse(docs model.DocumentSlice) ([]byte, error) {
	var respData []interface{}
	for _, doc := range docs {
		respData = append(respData, doc.Data["id"])
	}
	return json.Marshal(map[string]interface{}{
		"statements": respData,
		"more":       "",
	})
}

func formatRespMultipleStatements(formatType string, languages acceptlang.AcceptLanguages, docs model.DocumentSlice) ([]byte, error) {
	switch formatType {
	case "exact":
		return formatExactResponse(docs)
	case "ids":
		return formatIDsResponse(docs)
	case "canonical":
		return formatCanonicalResponse(docs, languages)
	default:
		return nil, errors.New("invalid formatType given")
	}
}

func collectSHA2sOfAttachments(statments model.DocumentSlice) []string {
	var sha2s []string
	for _, stmt := range statments {
		if atts, ok := stmt.Data["attachments"]; ok {
			if attifces, ok := atts.([]interface{}); ok {
				for _, attifce := range attifces {
					if att, ok := attifce.(map[string]interface{}); ok {
						if sha2, ok := att["sha2"]; ok {
							sha2s = append(sha2s, sha2.(string))
						}
					}
				}
			}
		}
	}

	return sha2s
}

func appendAttachments(session *mgo.Session, respBody []byte, sha2s []string) (body []byte, boundary string, err error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	boundary = w.Boundary()
	part := make(textproto.MIMEHeader)
	part.Set("Content-Type", "application/json")
	pw, err := w.CreatePart(part)
	if err != nil {
		return nil, "", err
	}
	io.Copy(pw, bytes.NewReader(respBody))

	gfs := session.DB(miscs.GlobalConfig.MongoDB.DBName).GridFS("fs")
	for _, sha2 := range sha2s {
		gfsfile, err := gfs.Open(sha2)
		if err != nil {
			return nil, "", err
		}
		var metadata map[string]interface{}
		err = gfsfile.GetMeta(&metadata)
		if err != nil {
			return nil, "", err
		}

		part := make(textproto.MIMEHeader)
		part.Set("Content-Type", metadata["Content-Type"].(string))
		part.Set("Content-Transfer-Encoding", "binary")
		part.Set("X-Experience-API-Hash", sha2)
		pw, err := w.CreatePart(part)
		if err != nil {
			return nil, "", err
		}
		io.Copy(pw, gfsfile)
	}
	w.Close()
	body = buf.Bytes()

	return
}
