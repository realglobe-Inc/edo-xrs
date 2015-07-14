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

package model

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Document struct {
	ID        bson.ObjectId          `bson:"_id,omitempty"`
	Version   string                 `bson:"version"`
	User      string                 `bson:"user"`
	App       string                 `bson:"app"`
	Timestamp time.Time              `bson:"timestamp"`
	Data      map[string]interface{} `bson:"data"`
}

func NewDocument(version, user, app string, timestamp time.Time, body bson.M) *Document {
	return &Document{
		bson.NewObjectId(),
		version,
		user,
		app,
		timestamp,
		body,
	}
}

func (d *Document) InsertTo(col *mgo.Collection) error {
	return col.Insert(d)
}

// DocumentSlice represents the slice of documents.
type DocumentSlice []Document

func (d DocumentSlice) toInterfaceArray() []interface{} {
	ifaces := make([]interface{}, 0, len(d))

	for _, doc := range d {
		ifaces = append(ifaces, doc)
	}

	return ifaces
}

func (d DocumentSlice) Len() int {
	return len(d)
}

func (d DocumentSlice) Less(i, j int) bool {
	return d[i].Timestamp.Before(d[j].Timestamp)
}

func (d DocumentSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DocumentSlice) InsertTo(col *mgo.Collection) error {
	return col.Insert((d.toInterfaceArray())...)
}

func (d DocumentSlice) Map(f func(Document) Document) {
	for ind, doc := range d {
		d[ind] = f(doc)
	}
}

func (d DocumentSlice) ToDataArray() (resp []bson.M) {
	for _, doc := range d {
		resp = append(resp, doc.Data)
	}

	return
}
