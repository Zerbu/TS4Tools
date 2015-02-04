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
	"fmt"
	"github.com/Fogity/TS4Tools/keys"
)

var (
	sims4Version = Version{2, 1}
)

func Open(path string) (*Package, error) {
	var p Package
	err := p.readFile(path)
	if err != nil {
		return nil, err
	}
	p.loadResourceList()
	return &p, nil
}

func New() *Package {
	var p Package
	p.initFile(sims4Version)
	p.resources = make([]*Resource, 0)
	return &p
}

func (p *Package) Save() error {
	if p.file == nil {
		return fmt.Errorf("no file associated with package")
	}
	p.saveResourceList()
	return p.writeFile(p.file.Name())
}

func (p *Package) SaveAs(path string) error {
	p.saveResourceList()
	return p.writeFile(path)
}

func (p *Package) Close() {
	p.file.Close()
	p.file = nil
}

func (p *Package) AddResource(resource *Resource) {
	p.resources = append(p.resources, resource)
}

func (p *Package) ListResources(include *keys.Filter, resources map[keys.Key]*Resource) map[keys.Key]*Resource {
	if resources == nil {
		resources = make(map[keys.Key]*Resource)
	}

	for _, resource := range p.resources {
		if include != nil {
			if !include.Include(resource) {
				continue
			}
		}
		resources[resource.key] = resource
	}

	return resources
}
