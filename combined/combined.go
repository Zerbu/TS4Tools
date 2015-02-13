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

package combined

import (
	"bytes"
	"encoding/xml"
)

type Combined struct {
	XMLName xml.Name `xml:"combined"`
	Entries []Entry  `xml:"R"`
}

type Entry struct {
	Type      string     `xml:"n,attr"`
	Instances []Instance `xml:"I"`
}

type Instance struct {
	XMLName  xml.Name
	Name     string    `xml:"n,attr"`
	Type     string    `xml:"t,attr"`
	Tunables []Tunable `xml:",any"`
}

type Tunable struct {
	XMLName  xml.Name
	Name     string    `xml:"n,attr"`
	Type     string    `xml:"t,attr"`
	Tunables []Tunable `xml:",any"`
	Value    string    `xml:",chardata"`
}

func Read(b []byte) (*Combined, error) {
	decoder := xml.NewDecoder(bytes.NewReader(b))
	c := new(Combined)
	err := decoder.Decode(c)
	return c, err
}
