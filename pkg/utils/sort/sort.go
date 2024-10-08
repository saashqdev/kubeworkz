/*
Copyright 2022 Kubeworkz Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sort

import "strings"

// ParseSort sortName=creationTimestamp, sortOrder=asc
func ParseSort(name string, order string, sFunc string) (sortName, sortOrder, sortFunc string) {
	sortName = "metadata.name"
	sortOrder = "asc"
	sortFunc = "string"

	if name == "" {
		return
	}
	sortName = name

	if strings.EqualFold(order, "desc") {
		sortOrder = "desc"
	}

	if sFunc != "" {
		sortFunc = sFunc
	}

	return
}
