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
	"encoding/binary"
	"fmt"
	"github.com/Fogity/TS4Tools/hash"
)

const (
	alignment = 16
)

type writeContext struct {
	d               *Simdata
	objectTables    map[*schema]*dataTable
	schemas         []*schema
	tableInfoOffset int
	schemaOffset    int
	schemaOffsets   map[*schema]int
	nameOffset      int
	names           []string
}

type dataTable struct {
	buffer *bytes.Buffer
	schema *schema
	length int
	count  int
	offset int
	name   string
}

func (c *writeContext) writeSimdata() []byte {
	c.objectTables = make(map[*schema]*dataTable)
	c.schemas = make([]*schema, 0)
	c.schemaOffsets = make(map[*schema]int)
	c.names = make([]string, 0)

	for _, object := range c.d.objects {
		c.writeObject(object)
	}

	for _, table := range c.objectTables {
		pad(table)
	}

	c.calculateOffsets()

	b := new(bytes.Buffer)

	h := new(header)
	h.Identifier = identifier
	h.Version = version
	h.TableInfoOffset = int32(c.tableInfoOffset) - headerTableInfoOffset
	h.TableInfoCount = int32(len(c.objectTables))
	h.SchemaOffset = int32(c.schemaOffset) - headerSchemaOffset
	h.SchemaCount = int32(len(c.schemas))

	err := binary.Write(b, binary.LittleEndian, h)
	if err != nil {
		panic(err)
	}
	b.Write(make([]byte, 8))
	offset := int32(binary.Size(h)) + 8

	for _, table := range c.objectTables {
		info := new(tableInfo)
		if table.name == "" {
			info.NameOffset = null
			info.NameHash = nullHash
		} else {
			off, hash := c.addName(table.name)
			info.NameOffset = off - offset
			info.NameHash = hash
		}
		info.SchemaOffset = int32(c.schemaOffsets[table.schema]) - offset - tableInfoSchemaOffset
		info.DataType = uint32(dtObject)
		info.RowSize = table.schema.header.SchemaSize
		info.RowOffset = int32(table.offset) - offset - tableInfoRowOffset
		info.RowCount = uint32(table.count)
		err = binary.Write(b, binary.LittleEndian, info)
		if err != nil {
			panic(err)
		}
		offset += int32(binary.Size(info))
	}

	padding := alignment - offset%alignment
	if padding == alignment {
		padding = 0
	}
	for i := 0; i < int(padding); i++ {
		b.WriteByte(0)
	}
	offset += padding

	for _, table := range c.objectTables {
		data := table.buffer.Bytes()
		err = binary.Write(b, binary.LittleEndian, data)
		if err != nil {
			panic(err)
		}
		offset += int32(binary.Size(data))
	}

	for _, schema := range c.schemas {
		if schema.name != "" {
			off, _ := c.addName(schema.name)
			schema.header.NameOffset = off - offset
		}
		err = binary.Write(b, binary.LittleEndian, schema.header)
		if err != nil {
			panic(err)
		}
		offset += int32(binary.Size(schema.header))

		for _, column := range schema.columns {
			if column.name != "" {
				off, _ := c.addName(column.name)
				column.header.NameOffset = off - offset
			}
			err = binary.Write(b, binary.LittleEndian, column.header)
			if err != nil {
				panic(err)
			}
			offset += int32(binary.Size(column.header))
		}
	}

	for _, name := range c.names {
		err = binary.Write(b, binary.LittleEndian, []byte(name))
		if err != nil {
			panic(err)
		}
		b.WriteByte(0)
	}

	return b.Bytes()
}

func (c *writeContext) addName(name string) (int32, uint32) {
	c.names = append(c.names, name)
	offset := int32(c.nameOffset)
	c.nameOffset += len(name) + 1
	return offset, hash.Fnv32(name)
}

func (c *writeContext) calculateOffsets() {
	offset := binary.Size(header{}) + 8
	c.tableInfoOffset = offset
	offset += len(c.objectTables) * binary.Size(tableInfo{})
	if offset%alignment != 0 {
		offset = (offset/alignment + 1) * alignment
	}
	for _, table := range c.objectTables {
		table.offset = offset
		offset += table.length
	}
	c.schemaOffset = offset
	for _, schema := range c.schemas {
		c.schemaOffsets[schema] = offset
		offset += binary.Size(schema.header) + int(schema.header.ColumnCount)*binary.Size(schemaColumn{})
	}
	c.nameOffset = offset
}

func (c *writeContext) writeObject(object *object) {
	table, ok := c.objectTables[object.schema]
	if !ok {
		table = new(dataTable)
		c.objectTables[object.schema] = table
		table.buffer = new(bytes.Buffer)
		table.name = object.name
		table.schema = object.schema
		c.schemas = append(c.schemas, object.schema)
	}
	table.count++
	b := make([]byte, object.schema.header.SchemaSize)
	for _, column := range object.schema.columns {
		value, ok := object.values[column.name]
		if !ok {
			panic(fmt.Errorf("value for column '%v' not found", column.name))
		}
		buf := new(bytes.Buffer)
		switch column.header.DataType {
		case dtString8, dtObject:
			panic(fmt.Errorf("writing column data type (%v) not implemented", column.header.DataType))
		case dtHashedString8, dtVector:
			panic(fmt.Errorf("writing column data type (%v) not implemented", column.header.DataType))
		default:
			err := binary.Write(buf, binary.LittleEndian, value)
			if err != nil {
				panic(err)
			}
		}
		off := int(column.header.Offset)
		val := buf.Bytes()
		for i := 0; i < len(val); i++ {
			b[off+i] = val[i]
		}
	}
	err := binary.Write(table.buffer, binary.LittleEndian, b)
	if err != nil {
		panic(err)
	}
	table.length += binary.Size(b)
}

func pad(table *dataTable) {
	padding := alignment - table.length%alignment
	if padding == alignment {
		return
	}
	for i := 0; i < padding; i++ {
		table.buffer.WriteByte(0)
	}
	table.length += padding
}
