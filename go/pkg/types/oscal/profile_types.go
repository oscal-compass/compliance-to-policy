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

package oscal

type ProfileRoot struct {
	Profile `json:"profile"`
}

type ProfileMetadata struct {
	Title        string `json:"title"`
	LastModified string `json:"last-modified"`
	Version      string `json:"version"`
	OscalVersion string `json:"oscal-version"`
	Roles        []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"roles"`
	Parties []struct {
		UUID           string   `json:"uuid"`
		Type           string   `json:"type"`
		Name           string   `json:"name"`
		EmailAddresses []string `json:"email-addresses"`
		Addresses      []struct {
			AddrLines  []string `json:"addr-lines"`
			City       string   `json:"city"`
			State      string   `json:"state"`
			PostalCode string   `json:"postal-code"`
		} `json:"addresses"`
	} `json:"parties"`
	ResponsibleParties []struct {
		RoleID     string   `json:"role-id"`
		PartyUuids []string `json:"party-uuids"`
	} `json:"responsible-parties"`
}

type ProfileImport struct {
	Href            string `json:"href"`
	IncludeControls []struct {
		WithIds []string `json:"with-ids"`
	} `json:"include-controls"`
}

type Profile struct {
	UUID     string          `json:"uuid"`
	Metadata ProfileMetadata `json:"metadata"`
	Imports  []ProfileImport `json:"imports"`
	Merge    struct {
		AsIs bool `json:"as-is"`
	} `json:"merge"`
}
