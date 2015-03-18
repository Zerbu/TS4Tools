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

package script

import (
	"github.com/Fogity/TS4Tools/dbpf"
	"github.com/Fogity/TS4Tools/keys"
)

type Package struct {
	p       *dbpf.Package
	merge   []*dbpf.Package
	include *keys.Filter
	exclude *keys.Filter
}

func CreatePackage() *Package {
	p := &Package{
		p: dbpf.New(),
	}
	return p
}

func OpenPackage(path string) (*Package, error) {
	pack, err := dbpf.Open(path)
	p := &Package{
		p: pack,
	}
	return p, err
}

func (p *Package) Merge(path string) error {
	pack, err := dbpf.Open(path)
	if err != nil {
		return err
	}
	if p.merge == nil {
		p.merge = make([]*dbpf.Package, 0)
	}
	p.merge = append(p.merge, pack)
	return nil
}

func (p *Package) Include(filter *keys.Filter) {
	p.include = keys.MergeFilters(p.include, filter)
}

func (p *Package) Exclude(filter *keys.Filter) {
	p.exclude = keys.MergeFilters(p.exclude, filter)
}
