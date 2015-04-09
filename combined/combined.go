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
	"fmt"
)

type Combined struct {
	XMLName xml.Name `xml:"combined"`
	Entries []Entry  `xml:"R"`
}

type Entry struct {
	Type      string     `xml:"n,attr"`
	Instances []Instance `xml:"I"`
	Modules   []Instance `xml:"M"`
}

type Instance struct {
	XMLName  xml.Name
	Class    string    `xml:"c,attr"`
	Instance string    `xml:"i,attr"`
	Module   string    `xml:"m,attr"`
	Name     string    `xml:"n,attr"`
	Id       string    `xml:"s,attr"`
	Tunables []Tunable `xml:",any"`
}

type Tunable struct {
	XMLName   xml.Name
	Type      string    `xml:",attr"`
	Path      string    `xml:"p,attr"`
	Enum      string    `xml:"ev,attr"`
	Name      string    `xml:"n,attr"`
	Reference string    `xml:"x,attr"`
	Tunables  []Tunable `xml:",any"`
	Value     string    `xml:",chardata"`
}

func Read(b []byte) (*Combined, error) {
	decoder := xml.NewDecoder(bytes.NewReader(b))
	c := new(Combined)
	err := decoder.Decode(c)
	if err != nil {
		return nil, err
	}

	references := extractReferences(b)

	c = dereference(c, references)

	return c, nil
}

func dereference(combined *Combined, references map[int]Tunable) *Combined {
	var c Combined
	c = *combined
	for k, entry := range c.Entries {
		e := entry
		for m, instance := range e.Instances {
			i := instance
			for p, tunable := range i.Tunables {
				i.Tunables[p] = dereferenceCopy(tunable, references)
			}
			e.Instances[m] = i
		}
		for m, instance := range e.Modules {
			i := instance
			for p, tunable := range i.Tunables {
				i.Tunables[p] = dereferenceCopy(tunable, references)
			}
			e.Modules[m] = i
		}
		c.Entries[k] = e
	}
	return &c
}

func dereferenceCopy(tunable Tunable, references map[int]Tunable) Tunable {
	var t Tunable
	if tunable.XMLName.Local == "r" {
		var x int
		fmt.Sscan(tunable.Reference, &x)
		t = references[x]
		if tunable.Name != "" {
			t.Name = tunable.Name
		}
	} else {
		t = tunable
	}

	for i, st := range t.Tunables {
		t.Tunables[i] = dereferenceCopy(st, references)
	}

	return t
}

func extractReferences(b []byte) map[int]Tunable {
	decoder := xml.NewDecoder(bytes.NewReader(b))
	references := make(map[int]Tunable)
	for {
		token, _ := decoder.Token()

		if token == nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "r" {
				continue
			}
			for _, attr := range t.Attr {
				if attr.Name.Local != "x" {
					continue
				}
				var x int
				fmt.Sscan(attr.Value, &x)
				var tunable Tunable
				decoder.DecodeElement(&tunable, &t)
				references[x] = tunable
			}
		}
	}
	return references
}
