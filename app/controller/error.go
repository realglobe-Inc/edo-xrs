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
	"encoding/json"
	"fmt"
	"net/http"
)

type BadRequestErr struct {
	message string
}

func NewBadRequestErr(message string) BadRequestErr {
	return BadRequestErr{
		message: message,
	}
}

func NewBadRequestErrF(format string, a ...interface{}) BadRequestErr {
	return BadRequestErr{
		message: fmt.Sprintf(format, a...),
	}
}

func (e BadRequestErr) Response() (int, string) {
	return http.StatusBadRequest, e.JSON()
}

func (e BadRequestErr) JSON() string {
	body, err := json.Marshal(map[string]interface{}{
		"title":  "invalid request body",
		"status": 400,
		"detail": e.message,
	})
	if err != nil {
		logger.Err("Unexpected error occured:", err)
	}

	return string(body)
}

func (e BadRequestErr) String() string {
	return e.message
}
