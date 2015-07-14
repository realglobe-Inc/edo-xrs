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
	"gopkg.in/mgo.v2"
)

func InitDB(db *mgo.Database) {
	fatalOnErr(ensureUniqueIndexOn(db.C("quota"), []string{"user"}))
	fatalOnErr(ensureIndexOn(db.C("statement"), []string{"version", "user", "app"}))
	fatalOnErr(ensureUniqueIndexOn(db.C("statement"), []string{"version", "user", "app", "data.id"}))
}

func ensureIndexOn(coll *mgo.Collection, keys []string) error {
	return coll.EnsureIndex(mgo.Index{
		Key:        keys,
		Unique:     false,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	})
}

func ensureUniqueIndexOn(coll *mgo.Collection, keys []string) error {
	return coll.EnsureIndex(mgo.Index{
		Key:        keys,
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	})
}

func fatalOnErr(err error) {
	if err != nil {
		panic("An error occured on create indexes: " + err.Error())
	}
}
