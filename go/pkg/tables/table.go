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

package tables

import (
	"encoding/csv"
	"io"
	"strings"
)

type Row struct {
	Name        string
	Group       string
	Standard    string
	Category    string
	Control     string
	Source      string
	Destination string
}

type Table struct {
	rows []Row
}

func (table *Table) Add(row Row) {
	standards := strings.Split(row.Standard, ",")
	categories := strings.Split(row.Category, ",")
	controls := strings.Split(row.Control, ",")
	for i := 0; i < len(standards); i++ {
		_row := Row{
			Name:        row.Name,
			Group:       row.Group,
			Standard:    strings.TrimSpace(standards[i]),
			Category:    strings.TrimSpace(categories[i]),
			Control:     strings.TrimSpace(controls[i]),
			Source:      row.Source,
			Destination: row.Destination,
		}
		table.rows = append(table.rows, _row)
	}
}

func (table *Table) ToCsv(writer io.Writer) {
	wr := csv.NewWriter(writer)
	_ = wr.Write([]string{
		"name", "group", "standard", "category", "control", "source", "destination",
	})
	for _, row := range table.rows {
		_row := []string{
			row.Name, row.Group, row.Standard, row.Category, row.Control, row.Source, row.Destination,
		}
		_ = wr.Write(_row)
	}
	wr.Flush()
}
