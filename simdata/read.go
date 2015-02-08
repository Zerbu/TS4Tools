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
	"github.com/Fogity/TS4Tools/keys"
	"sort"
)

const (
	seekAbsolute = 0
)

type readContext struct {
	header  *header
	schemas map[int64]*schema
	tables  map[int64]*table
	data    map[int64]interface{}
	sizes   map[int64]int64
	r       *bytes.Reader
	p       int64
}

func (c *readContext) seek(offset int64) {
	p, err := c.r.Seek(offset, seekAbsolute)
	c.p = p
	if err != nil {
		panic(err)
	}
}

func (c *readContext) read(v interface{}) {
	err := binary.Read(c.r, binary.LittleEndian, v)
	c.p += int64(binary.Size(v))
	if err != nil {
		panic(err)
	}
}

func (c *readContext) readString() string {
	var b byte
	bs := make([]byte, 0)
	for {
		c.read(&b)
		if b == 0 {
			break
		}
		bs = append(bs, b)
	}
	c.p += int64(len(bs) + 1)
	return string(bs)
}

func (c *readContext) readSimdata() *Simdata {
	c.header = new(header)
	c.read(c.header)

	tableInfoOffset := c.p + int64(c.header.TableInfoOffset) - headerTableInfoAdjust
	schemaOffset := c.p + int64(c.header.SchemaOffset) - headerSchemaAdjust

	c.seek(schemaOffset)

	c.schemas = make(map[int64]*schema)
	for i := 0; i < int(c.header.SchemaCount); i++ {
		p := c.p
		schema := c.readSchema()
		c.schemas[p] = schema
	}

	c.seek(tableInfoOffset)

	c.tables = make(map[int64]*table)
	c.data = make(map[int64]interface{})
	c.sizes = make(map[int64]int64)
	for i := 0; i < int(c.header.TableInfoCount); i++ {
		p := c.p
		table := c.readTable()
		c.tables[p] = table
	}

	keys := make([]int, 0)
	for key := range c.tables {
		keys = append(keys, int(key))
	}
	sort.Ints(keys)

	for i := len(keys) - 1; i >= 0; i-- {
		table := c.tables[int64(keys[i])]
		if table.schema == nil {
			continue
		}
		offset := table.start
		for k := 0; k < int(table.header.RowCount); k++ {
			c.data[offset] = c.readObject(offset, table.schema, table.name)
			offset += int64(table.header.RowSize)
		}
	}

	objects := make(map[string]*object)
	for _, value := range c.data {
		object, ok := value.(*object)
		if !ok {
			continue
		}
		if object.name == "" {
			continue
		}
		objects[object.name] = object
	}

	return &Simdata{objects}
}

func (c *readContext) readSchema() *schema {
	h := new(schemaHeader)
	c.read(h)

	nameOffset := c.p + int64(h.NameOffset) - schemaHeaderNameAdjust
	columnOffset := c.p + int64(h.ColumnOffset) - schemaHeaderColumnAdjust

	p := c.p

	var name string
	if h.NameOffset != null {
		c.seek(nameOffset)
		name = c.readString()
	}

	c.seek(columnOffset)
	columns := make([]*column, h.ColumnCount)
	for i := range columns {
		columns[i] = c.readColumn()
	}

	c.seek(p)

	return &schema{h, columns, name}
}

func (c *readContext) readColumn() *column {
	col := new(schemaColumn)
	c.read(col)

	nameOffset := c.p + int64(col.NameOffset) - schemaColumnNameAdjust

	p := c.p

	var name string
	if col.NameOffset != null {
		c.seek(nameOffset)
		name = c.readString()
	}

	c.seek(p)

	return &column{col, name}
}

func (c *readContext) readTable() *table {
	h := new(tableInfo)
	c.read(h)

	nameOffset := c.p + int64(h.NameOffset) - tableInfoNameAdjust
	schemaOffset := c.p + int64(h.SchemaOffset) - tableInfoSchemaAdjust
	rowOffset := c.p + int64(h.RowOffset) - tableInfoRowAdjust

	p := c.p

	var name string
	if h.NameOffset != null {
		c.seek(nameOffset)
		name = c.readString()
	}

	if int(h.DataType) != dtObject {
		c.seek(rowOffset)
		for i := 0; i < int(h.RowCount); i++ {
			offset := c.p
			c.data[offset] = c.readValue(int(h.DataType))
			c.sizes[offset] = int64(h.RowSize)
		}
	}

	c.seek(p)

	return &table{h, c.schemas[schemaOffset], rowOffset, name}
}

func (c *readContext) readValue(dataType int) interface{} {
	switch dataType {
	case dtChar8:
		var char uint8
		c.read(&char)
		return char
	case dtInt32:
		var i int32
		c.read(&i)
		return i
	case dtInt64:
		var i int64
		c.read(&i)
		return i
	case dtFloat:
		var f float32
		c.read(&f)
		return f
	case dtString8:
		var offset uint32
		c.read(&offset)
		p := c.p
		c.seek(p + int64(offset) - 4)
		str := c.readString()
		c.seek(p)
		return str
	case dtVector:
		var offset, count uint32
		c.read(&offset)
		off := c.p + int64(offset) - 4
		c.read(&count)
		vector := make([]interface{}, count)
		for i := 0; i < int(count); i++ {
			val, ok := c.data[off]
			if !ok {
				panic(fmt.Errorf("element in vector not found"))
			}
			vector[i] = val
			off = off + c.sizes[off]
		}
		return vector
	case dtObject:
		var offset uint32
		c.read(&offset)
		off := c.p + int64(offset) - 4
		if int32(offset) != null {
			obj, ok := c.data[off]
			if !ok {
				panic(fmt.Errorf("object not found"))
			}
			return obj
		}
		return nil
	case dtFloat3:
		var fs [3]float32
		c.read(&fs)
		return fs
	case dtTableSetReference:
		var value uint64
		c.read(&value)
		return value
	case dtResourceKey:
		var t, g uint32
		var i uint64
		c.read(&i)
		c.read(&t)
		c.read(&g)
		return keys.Key{t, g, i}
	case dtLocKey:
		var key uint32
		c.read(&key)
		return key
	default:
		panic(fmt.Errorf("data type (%v) not implemented", dataType))
	}
}

func (c *readContext) readObject(offset int64, schema *schema, name string) *object {
	values := make(map[string]interface{})
	for _, column := range schema.columns {
		c.seek(offset + int64(column.header.Offset))
		values[column.name] = c.readValue(int(column.header.DataType))
	}
	return &object{schema, values, name}
}
