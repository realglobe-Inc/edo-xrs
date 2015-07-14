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

// XAPIVersion is for specify version of xAPI.
type XAPIVersion int

const (
	// XAPIVersionVoid indicates invalid XAPI version
	XAPIVersionVoid XAPIVersion = iota
	// XAPIVersion102 indicates xAPI version 1.0.2.
	XAPIVersion102
)

// SupportedVersions returns all versions supported by this validator.
func SupportedVersions() []XAPIVersion {
	return []XAPIVersion{XAPIVersion102}
}

// ToXAPIVersion converts XAPI version string to constant.
func ToXAPIVersion(version string) XAPIVersion {
	switch version {
	case "1.0.2":
		return XAPIVersion102
	default:
		return XAPIVersionVoid
	}
}

// IsValidXAPIVersion validates XAPI version string.
func IsValidXAPIVersion(version string) bool {
	return ToXAPIVersion(version) != XAPIVersionVoid
}
