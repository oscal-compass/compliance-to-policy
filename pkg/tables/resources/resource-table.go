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

package resources

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
)

var logger *zap.Logger = pkg.GetLogger("resource-tables")

type op int

const (
	get op = iota
	set
)

type Row struct {
	Kind         string
	ApiVersion   string
	Name         string
	Policy       string
	ConfigPolicy string
	Standard     string
	Category     string
	Control      string
	Source       string
	PolicyDir    string
}

type atomicOperateOut struct {
	ok    bool
	value string
}

func atomicOperate(op op, value *string, args ...string) atomicOperateOut {
	if op == set {
		if len(args) == 0 {
			return atomicOperateOut{ok: false}
		}
		*value = args[0]
		return atomicOperateOut{ok: true}
	} else if op == get {
		return atomicOperateOut{ok: true, value: *value}
	}
	return atomicOperateOut{ok: false}
}

func (row *Row) Get(column string) string {
	return row.get(column)
}

func (row *Row) get(column string) string {
	atomicOperateOut := row.access(column, get, "")
	if atomicOperateOut.ok {
		return atomicOperateOut.value
	} else {
		return ""
	}
}

func (row *Row) set(column string, value string) bool {
	atomicOperateOut := row.access(column, set, value)
	return atomicOperateOut.ok
}

func (row *Row) access(column string, op op, value string) atomicOperateOut {
	switch column {
	case "kind":
		return atomicOperate(op, &row.Kind, value)
	case "api-version":
		return atomicOperate(op, &row.ApiVersion, value)
	case "name":
		return atomicOperate(op, &row.Name, value)
	case "policy":
		return atomicOperate(op, &row.Policy, value)
	case "config-policy":
		return atomicOperate(op, &row.ConfigPolicy, value)
	case "standard":
		return atomicOperate(op, &row.Standard, value)
	case "category":
		return atomicOperate(op, &row.Category, value)
	case "control":
		return atomicOperate(op, &row.Control, value)
	case "source":
		return atomicOperate(op, &row.Source, value)
	case "policy-dir":
		return atomicOperate(op, &row.PolicyDir, value)
	}
	return atomicOperateOut{ok: false}
}

var columns = []string{
	"kind",
	"api-version",
	"name",
	"policy",
	"config-policy",
	"standard",
	"category",
	"control",
	"source",
	"policy-dir",
}

func GetColumns() []string {
	return columns
}

func FromCsv(reader io.Reader) *Table {
	t := Table{}
	csvReader := csv.NewReader(reader)
	data, err := csvReader.ReadAll()
	if err != nil {
		logger.Sugar().Errorf("%v", err)
	}
	var indexToColumn []string
	for i, line := range data {
		if i == 0 {
			indexToColumn = append(indexToColumn, line...)
		}
		if i > 0 { // omit header line
			var row Row
			for j, field := range line {
				column := indexToColumn[j]
				row.set(column, field)
			}
			t.Add(row)
		}
	}
	return &t
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
			Name:         row.Name,
			Kind:         row.Kind,
			ApiVersion:   row.ApiVersion,
			Policy:       row.Policy,
			ConfigPolicy: row.ConfigPolicy,
			Standard:     strings.TrimSpace(standards[i]),
			Category:     strings.TrimSpace(categories[i]),
			Control:      strings.TrimSpace(controls[i]),
			Source:       row.Source,
			PolicyDir:    row.PolicyDir,
		}
		table.rows = append(table.rows, _row)
	}
}

func (table *Table) List() []Row {
	return table.rows
}

func (table *Table) Filter(predicate func(row Row) bool) *Table {
	var filtered Table
	for _, row := range table.rows {
		if predicate(row) {
			filtered.Add(row)
		}
	}
	return &filtered
}

func (table *Table) GroupBy(column string) map[string]*Table {
	groupedTables := map[string]*Table{}
	for _, row := range table.rows {
		group := row.get(column)
		_, ok := groupedTables[group]
		if ok {
			groupedTables[group].Add(row)
		} else {
			table := Table{}
			table.Add(row)
			groupedTables[group] = &table
		}
	}
	return groupedTables
}

func (table *Table) ToCsv(writer io.Writer) {
	wr := csv.NewWriter(writer)
	_ = wr.Write(columns)
	for _, row := range table.rows {
		_row := []string{
			row.Kind, row.ApiVersion, row.Name, row.Policy, row.ConfigPolicy, row.Standard, row.Category, row.Control, row.Source, row.PolicyDir,
		}
		_ = wr.Write(_row)
	}
	wr.Flush()
}

func (table *Table) Print() {
	data := [][]string{}
	for _, row := range table.rows {
		dataRow := []string{}
		for _, column := range columns {
			dataRow = append(dataRow, row.get(column))
		}
		data = append(data, dataRow)
	}
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(columns)
	tw.SetBorder(false)
	tw.AppendBulk(data)
	tw.Render()
}
