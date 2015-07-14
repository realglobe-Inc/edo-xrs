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
	"github.com/realglobe-Inc/go-lib/rglog"
	"gopkg.in/mgo.v2"
)

var (
	logger = rglog.Logger("LRS/statement")
)

// Controller stores the local configuration of controller.
type Controller struct {
	session *mgo.Session // Mongo DB のセッション
}

// New は新しい Controller のインスタンスを返す。
// 引数の s にはこのコントローラで使用する Mongo DB のセッションが与えられる。
func New(s *mgo.Session) *Controller {
	return &Controller{s}
}
