/*
Copyright 2023 Kubeworkz Authors

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

package resourcemanage

import (
	"net/http"

	"github.com/saashqdev/kubeworkz/pkg/utils/filter"
)

type ResponseFilter struct {
	Condition        *filter.Condition
	ConverterContext *filter.ConverterContext
}

func (f *ResponseFilter) filterResponse(r *http.Response) error {
	return filter.NewFilter(f.ConverterContext).ModifyResponse(r, f.Condition)
}
