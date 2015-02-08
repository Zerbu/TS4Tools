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

type Simdata struct {
	objects map[string]*object
}

type schema struct {
	header  *schemaHeader
	columns []*column
	name    string
}

type column struct {
	header *schemaColumn
	name   string
}

type table struct {
	header *tableInfo
	schema *schema
	start  int64
	name   string
}

type object struct {
	schema *schema
	values map[string]interface{}
	name   string
}

func Read(b []byte) (d *Simdata, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(error)
		}
	}()

	r := bytes.NewReader(b)

	c := new(readContext)

	c.r = r

	d = c.readSimdata()

	if len(d.objects) != 1 {
		e = fmt.Errorf("simdata with %v root objects not supported", len(d.objects))
	}

	return
}

func (d *Simdata) Write() (b []byte, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(error)
		}
	}()

	panic(fmt.Errorf("writing not implemented"))

	return
}

func (d *Simdata) GetValue(name string) (interface{}, bool) {
	for _, object := range d.objects {
		val, ok := object.values[name]
		return val, ok
	}
	return nil, false
}

func (d *Simdata) SetValue(name string, value interface{}) error {
	for _, object := range d.objects {
		found := false
		for _, column := range object.schema.columns {
			if column.name == name {
				ok, err := checkType(value, int(column.header.DataType))
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("value not of correct type")
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("object do not have the specified field")
		}
		object.values[name] = value
		return nil
	}
	return fmt.Errorf("no object found")
}

func checkType(value interface{}, dataType int) (bool, error) {
	switch dataType {
	case dtFloat:
		_, ok := value.(float32)
		return ok, nil
	default:
		return false, fmt.Errorf("data type (%v) not implemented", dataType)
	}
}
