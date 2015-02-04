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

package dbpf

import (
	"os"
)

type Package struct {
	header    header
	record    record
	file      *os.File
	resources []*Resource
}

type Version struct {
	Major, Minor uint32
}

type header struct {
	Identifier        [4]byte
	FileVersion       Version
	UserVersion       Version
	_                 uint32
	CreationTime      uint32
	UpdateTime        uint32
	_                 uint32
	EntryCount        uint32
	RecordPositionLow uint32
	RecordSize        uint32
	_, _, _           uint32
	Unused            uint32
	RecordPosition    uint64
	_, _, _, _, _, _  uint32
}

type record struct {
	Flags      uint32
	Type       uint32
	Group      uint32
	InstanceEx uint32
	Entries    []*entry
}

type entry struct {
	Type       uint32
	Group      uint32
	InstanceEx uint32
	Fixed      entryFixed
	Extended   entryExtended
}

type entryFixed struct {
	Instance         uint32
	Position         uint32
	CompressedSize   uint32
	DecompressedSize uint32
}

type entryExtended struct {
	CompressionType uint16
	Committed       uint16
}
