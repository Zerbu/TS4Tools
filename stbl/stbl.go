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

package stbl

import (
	"bytes"
	"encoding/binary"
)

const (
	sims4version = 5
)

type header struct {
	Identifier   [4]byte
	Version      uint16
	Compressed   byte
	NumEntries   uint64
	_            [2]byte
	StringLength uint32
}

type Table struct {
	header  header
	Entries []Entry
}

type Entry struct {
	Key    uint32
	Flags  byte
	Length uint16
	String string
}

func Read(data []byte) (*Table, error) {
	table := new(Table)
	r := bytes.NewReader(data)

	err := binary.Read(r, binary.LittleEndian, &table.header)
	if err != nil {
		return nil, err
	}

	table.Entries = make([]Entry, table.header.NumEntries)
	for i := range table.Entries {
		var entry Entry
		binary.Read(r, binary.LittleEndian, &entry.Key)
		binary.Read(r, binary.LittleEndian, &entry.Flags)
		binary.Read(r, binary.LittleEndian, &entry.Length)
		array := make([]byte, entry.Length)
		r.Read(array)
		entry.String = string(array)
		table.Entries[i] = entry
	}

	return table, nil
}
