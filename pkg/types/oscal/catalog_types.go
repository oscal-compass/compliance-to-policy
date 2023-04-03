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

type CatalogRoot struct {
	Catalog `json:"catalog"`
}

type CatalogMetadata struct {
	Title        string `json:"title"`
	LastModified string `json:"last-modified"`
	Version      string `json:"version"`
	OscalVersion string `json:"oscal-version"`
	Props        []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"props"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Roles []struct {
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

type CatalogResource struct {
	UUID     string `json:"uuid"`
	Title    string `json:"title,omitempty"`
	Citation struct {
		Text string `json:"text,omitempty"`
	} `json:"citation,omitempty"`
	Rlinks []struct {
		Href string `json:"href,omitempty"`
	} `json:"rlinks,omitempty"`
}
type Catalog struct {
	UUID       string          `json:"uuid"`
	Metadata   CatalogMetadata `json:"metadata,omitempty"`
	Groups     []Group         `json:"groups,omitempty"`
	BackMatter struct {
		Resources []CatalogResource `json:"resources,omitempty"`
	} `json:"back-matter,omitempty"`
}

type Group struct {
	ID       string    `json:"id"`
	Class    string    `json:"class"`
	Title    string    `json:"title"`
	Controls []Control `json:"controls"`
	Groups   []Group   `json:"groups"`
	Parts    []struct {
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Title string `json:"title,omitempty"`
		Prose string `json:"prose,omitempty"`
	} `json:"parts,omitempty"`
}

type Control struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	Title  string `json:"title"`
	Params []struct {
		ID    string `json:"id"`
		Props []struct {
			Name  string `json:"name,omitempty"`
			Ns    string `json:"ns,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"props,omitempty"`
		Label      string `json:"label,omitempty"`
		Guidelines []struct {
			Prose string `json:"prose,omitempty"`
		} `json:"guidelines,omitempty"`
		Select struct {
			HowMany string   `json:"how-many,omitempty"`
			Choice  []string `json:"choice,omitempty"`
		} `json:"select,omitempty"`
	} `json:"params,omitempty"`
	Props []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		Class string `json:"class,omitempty"`
	} `json:"props,omitempty"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links,omitempty"`
	Parts []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Parts []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Props []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"props"`
			Prose string `json:"prose"`
			Parts []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Props []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"props"`
				Prose string `json:"prose"`
				Parts []struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Props []struct {
						Name  string `json:"name"`
						Value string `json:"value"`
					} `json:"props"`
					Prose string `json:"prose"`
				} `json:"parts,omitempty"`
			} `json:"parts,omitempty"`
		} `json:"parts,omitempty"`
		Prose string `json:"prose,omitempty"`
		Props []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
			Class string `json:"class"`
		} `json:"props,omitempty"`
	} `json:"parts,omitempty"`
	Controls []struct {
		ID     string `json:"id"`
		Class  string `json:"class"`
		Title  string `json:"title"`
		Params []struct {
			ID    string `json:"id"`
			Props []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
				Class string `json:"class,omitempty"`
			} `json:"props"`
			Label      string `json:"label"`
			Guidelines []struct {
				Prose string `json:"prose"`
			} `json:"guidelines"`
		} `json:"params,omitempty"`
		Props []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
			Class string `json:"class,omitempty"`
		} `json:"props"`
		Links []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Parts []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Prose string `json:"prose,omitempty"`
			Props []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
				Class string `json:"class"`
			} `json:"props,omitempty"`
			Parts []struct {
				Name  string `json:"name"`
				Prose string `json:"prose"`
			} `json:"parts,omitempty"`
		} `json:"parts,omitempty"`
	} `json:"controls,omitempty"`
}
