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
	"github.com/Fogity/TS4Tools/hash"
)

var (
	identifier = [4]byte{'D', 'A', 'T', 'A'}
	nullHash   = hash.Fnv32("")
)

const (
	version = 0x100
	null    = int32(-0x7FFFFFFF) - 1
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

const (
	headerTableInfoOffset = int32(8)
	headerSchemaOffset    = int32(16)
)

type header struct {
	Identifier      [4]byte
	Version         uint32
	TableInfoOffset int32
	TableInfoCount  int32
	SchemaOffset    int32
	SchemaCount     int32
}

const (
	headerTableInfoAdjust = int64(16)
	headerSchemaAdjust    = int64(8)
)

const (
	tableInfoSchemaOffset = int32(8)
	tableInfoRowOffset    = int32(20)
)

type tableInfo struct {
	NameOffset   int32
	NameHash     uint32
	SchemaOffset int32
	DataType     uint32
	RowSize      uint32
	RowOffset    int32
	RowCount     uint32
}

const (
	tableInfoNameAdjust   = int64(28)
	tableInfoSchemaAdjust = int64(20)
	tableInfoRowAdjust    = int64(8)
)

const (
	schemaHeaderColumnOffset = int32(16)
)

type schemaHeader struct {
	NameOffset   int32
	NameHash     uint32
	SchemaHash   uint32
	SchemaSize   uint32
	ColumnOffset int32
	ColumnCount  uint32
}

const (
	schemaHeaderNameAdjust   = int64(24)
	schemaHeaderColumnAdjust = int64(8)
)

const (
	schemaColumnSchemaOffset = int32(24)
)

type schemaColumn struct {
	NameOffset   int32
	NameHash     uint32
	DataType     uint16
	Flags        uint16
	Offset       uint32
	SchemaOffset int32
}

const (
	schemaColumnNameAdjust   = int64(20)
	schemaColumnSchemaAdjust = int64(4)
)
