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
	"github.com/Fogity/TS4Tools/keys"
)

const (
	dtBool = iota
	dtChar8
	dtInt8
	dtUInt8
	dtInt16
	dtUInt16
	dtInt32
	dtUInt32
	dtInt64
	dtUInt64
	dtFloat
	dtString8
	dtHashedString8
	dtObject
	dtVector
	dtFloat2
	dtFloat3
	dtFloat4
	dtTableSetReference
	dtResourceKey
	dtLocKey
	dtUndefined
)

func readData(r *bytes.Reader, table *table, tables map[int64]*table) interface{} {
	switch table.info.DataType {
	case dtObject:
		if table.schema == nil {
			panic(fmt.Errorf("table schema not found"))
		}
		return readSchemaData(r, table, tables)
	default:
		list := make([]interface{}, table.info.RowCount)
		for i := range list {
			list[i] = readValue(r, int(table.info.DataType), tables)
		}
		return list
	}
}

func readSchemaData(r *bytes.Reader, table *table, tables map[int64]*table) map[string]interface{} {
	data := make(map[string]interface{})
	start := locate(r)

	for _, column := range table.schema.columns {
		seek(r, start+int64(column.column.Offset))
		data[column.name] = readValue(r, int(column.column.DataType), tables)
	}

	return data
}

func readValue(r *bytes.Reader, dataType int, tables map[int64]*table) interface{} {
	switch dataType {
	case dtBool:
		var value bool
		read(r, &value)
		return value

	case dtChar8:
		var value uint8
		read(r, &value)
		return value

	case dtInt32:
		var value int32
		read(r, &value)
		return value

	case dtInt64:
		var value int64
		read(r, &value)
		return value

	case dtFloat:
		var value float32
		read(r, &value)
		return value

	case dtString8:
		fmt.Printf("Warning: data type 'string8' ignored\n")
		return nil

		var offset uint32
		read(r, &offset)
		if int32(offset) == null {
			return nil
		}
		off := absolute(r, int32(offset))
		curr := locate(r)
		seek(r, off)
		str := readString(r)
		seek(r, curr)
		return str

	case dtObject:
		fmt.Printf("Warning: data type 'object' ignored\n")
		return nil

	case dtVector:
		var offset, count uint32
		read(r, &offset)
		if int32(offset) == null {
			return nil
		}
		off := absolute(r, int32(offset))
		dt, ok := findDataType(off, tables)
		if !ok {
			panic(fmt.Errorf("could not find data type of vector elements"))
		}
		read(r, &count)
		array := make([]interface{}, count)
		curr := locate(r)
		seek(r, off)
		for i := range array {
			array[i] = readValue(r, dt, tables)
		}
		seek(r, curr)
		return array

	case dtFloat3:
		var value [3]float32
		read(r, &value)
		return value

	case dtTableSetReference:
		var value uint64
		read(r, &value)
		return value

	case dtResourceKey:
		var i uint64
		var t, g uint32
		read(r, &i)
		read(r, &t)
		read(r, &g)
		return keys.Key{t, g, i}

	case dtLocKey:
		var value uint32
		read(r, &value)
		return value

	default:
		panic(fmt.Errorf("data type '%v' not implemented", dataType))
	}
}

func findDataType(offset int64, tables map[int64]*table) (int, bool) {
	for _, table := range tables {
		count := int64(table.info.RowCount)
		size := int64(table.info.RowSize)
		for i := int64(0); i < count; i++ {
			off := table.offset + i*size
			if off == offset {
				return int(table.info.DataType), true
			}
		}
	}
	return dtUndefined, false
}
