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

package engine

import (
	"fmt"
	"github.com/Fogity/TS4Tools/dbpf"
	"github.com/Fogity/TS4Tools/keys"
	"github.com/Fogity/TS4Tools/simdata"
	"strings"
)

type session struct {
	parent   *session
	vars     map[string]interface{}
	merges   map[*dbpf.Package][]*dbpf.Package
	includes map[*dbpf.Package]*keys.Filter
}

func newSession() *session {
	s := new(session)
	s.vars = make(map[string]interface{})
	s.merges = make(map[*dbpf.Package][]*dbpf.Package)
	s.includes = make(map[*dbpf.Package]*keys.Filter)
	return s
}

func (s *session) close() {
	if s.parent != nil {
		s.transfer(s.parent)
	}
	for _, variable := range s.vars {
		switch v := variable.(type) {
		case *dbpf.Package:
			v.Close()
		}
	}
	for _, list := range s.merges {
		for _, p := range list {
			p.Close()
		}
	}
}

func (s *session) transfer(session *session) {
	for p, merges := range s.merges {
		list, ok := session.merges[p]
		if ok {
			for _, m := range merges {
				list = append(list, m)
			}
		} else {
			list = merges
		}
		session.merges[p] = list
	}
	for p, include := range s.includes {
		session.includes[p] = keys.MergeFilters(session.includes[p], include)
	}
}

func (s *session) panic(message string, args ...interface{}) {
	panic(fmt.Errorf(message, args...))
}

func (s *session) introduce(name string, value interface{}) {
	_, ok := s.vars[name]
	if ok {
		s.panic("variable '%v' already defined", name)
	}
	s.vars[name] = value
}

func (s *session) fetch(name string) interface{} {
	parts := strings.Split(name, ".")
	v, ok := s.vars[parts[0]]
	if !ok {
		if s.parent != nil {
			v = s.parent.fetch(parts[0])
		} else {
			s.panic("variable '%v' not defined", parts[0])
		}
	}
	if len(parts) > 1 {
		return s.fetchAttribute(v, parts[1:])
	}
	return v
}

func (s *session) fetchAttribute(variable interface{}, attributes []string) interface{} {
	var attr interface{}
	var err error
	switch v := variable.(type) {
	case *dbpf.Resource:
		switch attributes[0] {
		case "key":
			attr = v.Key()
		default:
			s.panic("type resource does not have the attribute '%v'", attributes[0])
		}
	case *simdata.Simdata:
		attr, err = v.GetVariable(attributes[0])
		if err != nil {
			return nil
		}
	default:
		s.panic("variable type does not have attributes")
	}
	if len(attributes) > 1 {
		return s.fetchAttribute(attr, attributes[1:])
	}
	return attr
}

func (s *session) set(name string, value interface{}) {
	if !strings.Contains(name, ".") {
		s.vars[name] = value
		return
	}
	parts := strings.Split(name, ".")
	v, ok := s.vars[parts[0]]
	if !ok {
		s.panic("variable '%v' not defined", parts[0])
	}
	s.setAttribute(v, parts[1:], value)
}

func (s *session) setAttribute(variable interface{}, parts []string, value interface{}) {
	if len(parts) > 1 {
		v := s.fetchAttribute(variable, parts[:1])
		s.setAttribute(v, parts[1:], value)
		return
	}
	switch v := variable.(type) {
	case *dbpf.Resource:
		switch parts[0] {
		case "key":
			v.SetKey(value.(keys.Key))
		default:
			s.panic("type resource does not have the attribute '%v'", parts[0])
		}
	default:
		s.panic("variable type does not have attributes")
	}
}

func (s *session) newResource(kind string) *dbpf.Resource {
	switch kind {
	case "empty":
		return new(dbpf.Resource)
	default:
		s.panic("resource type '%v' not recognized", kind)
		return nil
	}
}

func (s *session) parseResource(resource *dbpf.Resource, kind string) interface{} {
	switch kind {
	case "simdata":
		bytes, err := resource.ToBytes()
		if err != nil {
			s.panic(err.Error())
		}
		data, err := simdata.Read(bytes)
		if err != nil {
			s.panic(err.Error())
		}
		return data
	default:
		s.panic("resource type '%v' not recognized", kind)
		return nil
	}
}

func (s *session) mergePackage(full, merge *dbpf.Package) {
	list := s.merges[full]
	if list == nil {
		list = make([]*dbpf.Package, 0)
	}
	s.merges[full] = append(list, merge)
}

func (s *session) mergeInclude(filter *keys.Filter, pack *dbpf.Package) {
	f := s.includes[pack]
	s.includes[pack] = keys.MergeFilters(f, filter)
}

func (s *session) list(list interface{}) []interface{} {
	switch l := list.(type) {
	case nil:
		return nil
	case *dbpf.Package:
		return s.listResources(l)
	case []uint64:
		il := make([]interface{}, len(l))
		for i, v := range l {
			il[i] = v
		}
		return il
	default:
		s.panic("variable type does not contain a list")
		return nil
	}
}

func (s *session) listResources(pack *dbpf.Package) []interface{} {
	resources := pack.ListResources(s.includes[pack], nil)
	for _, p := range s.merges[pack] {
		p.ListResources(s.includes[pack], resources)
	}
	list := make([]interface{}, 0)
	for _, resource := range resources {
		list = append(list, resource)
	}
	return list
}
