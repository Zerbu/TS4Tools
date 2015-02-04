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

var (
	identifier = [4]byte{'D', 'A', 'T', 'A'}
	version    = 0x100
	null       = int32(-0x7FFFFFFF) - 1
)

type Simdata struct {
	header  *header
	tables  []*table
	schemas []*schema
}

type table struct {
	name   string
	info   *tableInfo
	schema *schema
	data   map[string]interface{}
}

type schema struct {
	name    string
	header  *schemaHeader
	columns []*column
}

type column struct {
	name   string
	column *schemaColumn
}

func Read(b []byte) (d *Simdata, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	r := bytes.NewReader(b)
	d = readSimdata(r)

	return
}

func (d *Simdata) GetVariable(name string) (interface{}, error) {
	variable, ok := d.tables[0].data[name]
	if !ok {
		return nil, fmt.Errorf("could not find value")
	}
	return variable, nil
}
