/*
Copyright 2015 Henrik Rostedt <https://github.com/Fogity/>

This file is part of TS4Tools.

TS4Tools is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

TS4Tools is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with TS4Tools.  If not, see <http://www.gnu.org/licenses/>.
*/

package simdata

import (
	"bytes"
	"fmt"
)

func readData(r *bytes.Reader, info *tableInfo, schema *schema) interface{} {
	switch info.DataType {
	case dtObject:
		if schema == nil {
			panic(fmt.Errorf("table schema not found"))
		}
		return readSchemaData(r, schema)
	default:
		panic(fmt.Errorf("table data type not implemented (%v)", info.DataType))
	}
}

func readSchemaData(r *bytes.Reader, schema *schema) map[string]interface{} {
	data := make(map[string]interface{})
	start := locate(r)

	for _, column := range schema.columns {
		seek(r, start+int64(column.column.Offset))

		data[column.name] = readValue(r, int(column.column.DataType))
	}

	return data
}

func readValue(r *bytes.Reader, dataType int) interface{} {
	switch dataType {
	default:
		panic(fmt.Errorf("data type '%v' not implemented"))
	}
}
