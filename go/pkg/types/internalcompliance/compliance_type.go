/*
Copyright 2023 IBM Corporation

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

package internalcompliance

type Compliance struct {
	Standard Standard `json:"standard"`
}

type Standard struct {
	Name       string     `json:"name"`
	Categories []Category `json:"categories"`
}
type Category struct {
	Name     string    `json:"name"`
	Controls []Control `json:"controls"`
}
type Control struct {
	Name        string   `json:"name"`
	ControlRefs []string `json:"controlRefs"`
}
