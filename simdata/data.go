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

func readData(r *bytes.Reader, schema *schema) map[string]interface{} {
	data := make(map[string]interface{})
	start := locate(r)

	for _, column := range schema.columns {
		seek(r, start+int64(column.column.Offset))

		switch column.column.DataType {
		case dtInt64:
			var num int64
			read(r, &num)
			data[column.name] = num

		case dtFloat:
			var num float32
			read(r, &num)
			data[column.name] = num

		case dtObject:
			fmt.Printf("type 'Object' not implemented\n")

		case dtVector:
			var offset, count uint32
			read(r, &offset)
			off := absolute(r, int32(offset))
			read(r, &count)

			if int32(offset) == null {
				continue
			}

			seek(r, off)
			vector := make([]uint64, count)
			for i := range vector {
				var entry uint64
				read(r, &entry)
				vector[i] = entry
			}
			data[column.name] = vector

		case dtTableSetReference:
			fmt.Printf("type 'TableSetReference' not implemented\n")

		case dtResourceKey:
			fmt.Printf("type 'ResourceKey' not implemented\n")

		default:
			panic(fmt.Errorf("data type '%v' not implemented", column.column.DataType))
		}
	}

	return data
}
