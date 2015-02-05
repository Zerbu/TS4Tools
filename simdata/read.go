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
)

const (
	seekAbs = iota
	seekRel
	seekEnd
)

func locate(r *bytes.Reader) int64 {
	loc, err := r.Seek(0, seekRel)
	if err != nil {
		panic(err)
	}
	return loc
}

func seek(r *bytes.Reader, offset int64) {
	_, err := r.Seek(offset, seekAbs)
	if err != nil {
		panic(err)
	}
}

func absolute(r *bytes.Reader, offset int32) int64 {
	return locate(r) + int64(offset) - 4
}

func read(r *bytes.Reader, v interface{}) {
	err := binary.Read(r, binary.LittleEndian, v)
	if err != nil {
		panic(err)
	}
}

func readString(r *bytes.Reader) string {
	var b byte
	bs := make([]byte, 0)
	for {
		read(r, &b)
		if b == 0 {
			break
		}
		bs = append(bs, b)
	}
	return string(bs)
}

func readSimdata(r *bytes.Reader) *Simdata {
	h := new(header)
	read(r, &h.Identifier)
	read(r, &h.Version)
	read(r, &h.TableInfoOffset)
	tableInfoOffset := absolute(r, h.TableInfoOffset)
	read(r, &h.TableInfoCount)
	read(r, &h.SchemaOffset)
	schemaOffset := absolute(r, h.SchemaOffset)
	read(r, &h.SchemaCount)

	seek(r, schemaOffset)
	schemas := make(map[int64]*schema)
	for i := 0; i < int(h.SchemaCount); i++ {
		offset := locate(r)
		schemas[offset] = readSchema(r)
	}

	seek(r, tableInfoOffset)
	tables := make(map[int64]*table)
	for i := range tables {
		offset := locate(r)
		tables[offset] = readTable(r, schemas)
	}

	return &Simdata{h, tables, schemas}
}

func readSchema(r *bytes.Reader) *schema {
	n, name := readName(r)

	h := new(schemaHeader)
	h.Name = n
	read(r, &h.SchemaHash)
	read(r, &h.SchemaSize)
	read(r, &h.ColumnOffset)
	columnOffset := absolute(r, h.ColumnOffset)
	read(r, &h.ColumnCount)

	curr := locate(r)
	seek(r, columnOffset)
	columns := make([]*column, h.ColumnCount)
	for i := range columns {
		columns[i] = readColumn(r)
	}
	seek(r, curr)

	return &schema{name, h, columns}
}

func readColumn(r *bytes.Reader) *column {
	n, name := readName(r)

	c := new(schemaColumn)
	c.Name = n
	read(r, &c.DataType)
	read(r, &c.Flags)
	read(r, &c.Offset)
	read(r, &c.SchemaOffset)

	return &column{name, c}
}

func readTable(r *bytes.Reader, schemas map[int64]*schema) *table {
	n, name := readName(r)

	info := new(tableInfo)
	info.Name = n
	read(r, &info.SchemaOffset)
	schemaOffset := absolute(r, info.SchemaOffset)
	read(r, &info.DataType)
	read(r, &info.RowSize)
	read(r, &info.RowOffset)
	rowOffset := absolute(r, info.RowOffset)
	read(r, &info.RowCount)

	// TODO: read data

	return &table{name, info, schemas[schemaOffset], nil}
}

func readName(r *bytes.Reader) (name, string) {
	var n name

	read(r, &n.NameOffset)
	nameOffset := absolute(r, n.NameOffset)
	read(r, &n.NameHash)

	if n.NameOffset == null {
		return n, ""
	}

	curr := locate(r)
	seek(r, nameOffset)
	s := readString(r)
	seek(r, curr)

	return n, s
}
