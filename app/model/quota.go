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
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Quota struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	User  string        `bson:"user"`
	Usage int64         `bson:"usage"`
}

func (q *Quota) Check() bool {
	return q.Usage < miscs.GlobalConfig.Quota.UserMaxUsage
}

func (q *Quota) IncrementUsageTo(db *mgo.Database, amount int64) error {
	return db.C("quota").Update(bson.M{"_id": q.ID}, bson.M{"$inc": bson.M{"usage": amount}})
}

func GetQuota(db *mgo.Database, user string) (*Quota, error) {
	col := db.C("quota")

	var quota Quota
	if n, err := col.Find(bson.M{"user": user}).Count(); err == nil && n > 0 {
		if err = col.Find(bson.M{"user": user}).One(&quota); err != nil {
			return nil, err
		}

		return &quota, nil
	} else if n == 0 {
		quota.ID = bson.NewObjectId()
		quota.User = user
		quota.Usage = 0

		if err = col.Insert(quota); err != nil {
			return nil, err
		}

		return &quota, nil
	} else {
		return nil, err
	}

	return nil, nil // unreached
}
