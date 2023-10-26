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

package common

type Prop struct {
	Name    string `json:"name"`
	Ns      string `json:"ns,omitempty"`
	Value   string `json:"value"`
	Class   string `json:"class,omitempty"`
	Remarks string `json:"remarks,omitempty"`
}

type Link struct {
	Href             string `json:"href"`
	Rel              string `json:"rel,omitempty"`
	MediaType        string `json:"media-type,omitempty"`
	ResourceFragment string `json:"resource-fragment,omitempty"`
	Text             string `json:"text,omitempty"`
}

type RelevantEvidence struct {
	Href        string   `json:"href,omitempty"`
	Description string   `json:"description,omitempty"`
	Props       []Prop   `json:"props,omitempty"`
	Links       []Link   `json:"links,omitempty"`
	Remarks     []string `json:"remarks,omitempty"`
}
