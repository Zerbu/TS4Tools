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

type name struct {
	NameOffset int32
	NameHash   uint32
}

type header struct {
	Identifier      [4]byte
	Version         uint32
	TableInfoOffset int32
	TableInfoCount  int32
	SchemaOffset    int32
	SchemaCount     int32
}

type tableInfo struct {
	Name         name
	SchemaOffset int32
	DataType     uint32
	RowSize      uint32
	RowOffset    int32
	RowCount     uint32
}

type schemaHeader struct {
	Name         name
	SchemaHash   uint32
	SchemaSize   uint32
	ColumnOffset int32
	ColumnCount  uint32
}

type schemaColumn struct {
	Name         name
	DataType     uint16
	Flags        uint16
	Offset       uint32
	SchemaOffset int32
}
