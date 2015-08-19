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
	"regexp"

	"github.com/realglobe-Inc/gojsonschema"
)

var validUUID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

// IsUUID validates UUID
func IsUUID(text string) bool {
	return validUUID.MatchString(text)
}

func IsLangMap(langMap interface{}) bool {
	schema := GetXAPISchemaInstance()

	res, err := schema.Get(XAPIVersion10x, "langmap").Validate(gojsonschema.NewGoLoader(langMap))
	if err != nil {
		logger.Warn(err)
		return false
	}

	return res.Valid()
}
