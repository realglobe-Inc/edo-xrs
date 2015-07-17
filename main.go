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

package main

import (
	//	"net/http"
	//	_ "net/http/pprof"
	"flag"
	"net/http"
	"os"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/acceptlang"
	"github.com/realglobe-Inc/edo-xrs/app/controller"
	"github.com/realglobe-Inc/edo-xrs/app/miscs"
	"github.com/realglobe-Inc/edo-xrs/app/model"
	"github.com/realglobe-Inc/go-lib/rglog"
	"gopkg.in/mgo.v2"
)

var (
	logger = rglog.Logger("xRS/main")
)

var configFile string

func init() {
	var defaultConfig = "./conf/app.conf"
	if p := os.Getenv("GOPATH"); len(p) != 0 {
		defaultConfig = p + "/src/github.com/realglobe-Inc/edo-xrs/conf/app.conf"
	}
	flag.StringVar(&configFile, "config", defaultConfig, "path of config file")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logger.Err("config file not found: ", err)
		os.Exit(1)
	}

	miscs.InitConfig(configFile)
}

func main() {
	// start pprof (system profiler) server
	//go func() {
	//	logger.Info(http.ListenAndServe("localhost:6000", nil))
	//}()

	// start routing
	session, err := mgo.Dial(miscs.GlobalConfig.MongoDB.URL)
	defer session.Close()
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}
	model.InitDB(session.DB(miscs.GlobalConfig.MongoDB.DBName))

	c := controller.New(session)

	router := martini.Classic()
	router.Get("/", func() string { return "Welcome to xRS API Server." })

	// CROS support
	router.Options("**", func(params martini.Params, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-Experience-API-Version")
	})
	router.Put("/:user/:app/statements", c.StoreStatement)
	router.Post("/:user/:app/statements", c.StoreMultStatement)
	router.Get("/:user/:app/statements", acceptlang.Languages(), c.FindStatement)

	router.Run()
}
