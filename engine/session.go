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
	"github.com/Fogity/TS4Tools/caspart"
	"github.com/Fogity/TS4Tools/combined"
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
	excludes map[*dbpf.Package]*keys.Filter
}

func newSession() *session {
	s := new(session)
	s.vars = make(map[string]interface{})
	s.merges = make(map[*dbpf.Package][]*dbpf.Package)
	s.includes = make(map[*dbpf.Package]*keys.Filter)
	s.excludes = make(map[*dbpf.Package]*keys.Filter)
	s.vars["true"] = true
	s.vars["false"] = false
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
	for p, exclude := range s.excludes {
		session.excludes[p] = keys.MergeFilters(session.excludes[p], exclude)
	}
}

func (s *session) panic(message string, args ...interface{}) {
	panic(fmt.Errorf(message, args...))
}

func (s *session) is(value interface{}, kind string) bool {
	switch v := value.(type) {
	case keys.Key:
		return kind == "key"
	case *simdata.Simdata:
		return kind == "simdata"
	default:
		_ = v
		return false
	}
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
	var ok bool
	switch v := variable.(type) {
	case keys.Key:
		switch attributes[0] {
		case "type":
			attr = v.Type
		case "group":
			attr = v.Group
		case "instance":
			attr = v.Instance
		default:
			s.panic("type key does not have the attribute '%v'", attributes[0])
		}
	case *dbpf.Resource:
		switch attributes[0] {
		case "key":
			attr = v.Key()
		default:
			s.panic("type resource does not have the attribute '%v'", attributes[0])
		}
	case *simdata.Simdata:
		attr, ok = v.GetValue(attributes[0])
		if !ok {
			s.panic("could not find attribute '%v' in simdata", attributes[0])
		}
	case *combined.Combined:
		ok = false
		for _, entry := range v.Entries {
			if entry.Type == attributes[0] {
				list := make([]interface{}, 0)
				for _, instance := range entry.Instances {
					list = append(list, instance)
				}
				attr = list
				ok = true
				break
			}
		}
		if !ok {
			s.panic("could not find type %v in combined", attributes[0])
		}
	case combined.Instance:
		ok = false
		for _, tunable := range v.Tunables {
			if tunable.Name == attributes[0] {
				attr = tunable
				ok = true
				break
			}
		}
		if !ok {
			s.panic("could not find tunable %v in instance %v", attributes[0], v.Name)
		}
	case combined.Tunable:
		switch attributes[0] {
		case "name":
			attr = v.Name
		case "value":
			attr = v.Value
		default:
			s.panic("type tunable does not have the attribute '%v'", attributes[0])
		}
	case *caspart.CasPart:
		switch attributes[0] {
		case "name":
			return v.Name
		case "showInUI":
			return v.ParamFlags&caspart.ShowInUI != 0
		default:
			s.panic("type caspart does not have the attribute '%v'", attributes[0])
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
	v := s.fetch(parts[0])
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
	case *simdata.Simdata:
		err := v.SetValue(parts[0], value)
		if err != nil {
			s.panic(err.Error())
		}
	case *caspart.CasPart:
		switch parts[0] {
		case "name":
			v.Name = value.(string)
		case "showInUI":
			if value == true {
				v.ParamFlags |= caspart.ShowInUI
				return
			}
			if value == false {
				v.ParamFlags ^= caspart.ShowInUI
				return
			}
			s.panic("expected boolean")
		default:
			s.panic("type caspart does not have the attribute '%v'", parts[0])
		}
	default:
		s.panic("variable type does not have attributes")
	}
}

func (s *session) search(list, query interface{}) []interface{} {
	switch l := list.(type) {
	case combined.Tunable:
		q, ok := query.(string)
		if !ok {
			s.panic("expected string as query")
		}
		parts := strings.Split(q, ":")
		if len(parts) != 2 {
			s.panic("did not understand query %v", q)
		}
		switch parts[0] {
		case "tag":
			result := make([]interface{}, 0)
			var f func(combined.Tunable)
			f = func(tunable combined.Tunable) {
				for _, t := range tunable.Tunables {
					if t.XMLName.Local == parts[1] {
						result = append(result, t)
					}
					f(t)
				}
			}
			f(l)
			return result
		default:
			s.panic("not recognized attribute %v", parts[0])
			return nil
		}
	default:
		s.panic("variable type is not searchable")
		return nil
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
		if bytes == nil {
			return nil
		}
		data, err := simdata.Read(bytes)
		if err != nil {
			s.panic(err.Error())
		}
		return data
	case "combined":
		bytes, err := resource.ToBytes()
		if err != nil {
			s.panic(err.Error())
		}
		if bytes == nil {
			return nil
		}
		combined, err := combined.Read(bytes)
		if err != nil {
			s.panic(err.Error())
		}
		return combined
	case "caspart":
		bytes, err := resource.ToBytes()
		if err != nil {
			s.panic(err.Error())
		}
		caspart, err := caspart.Read(bytes)
		if err != nil {
			s.panic(err.Error())
		}
		return caspart
	default:
		s.panic("resource type '%v' not recognized", kind)
		return nil
	}
}

func (s *session) unparseResource(value interface{}) *dbpf.Resource {
	switch v := value.(type) {
	case *simdata.Simdata:
		bytes, err := v.Write()
		if err != nil {
			s.panic(err.Error())
		}
		resource := new(dbpf.Resource)
		resource.FromBytes(bytes)
		return resource
	case *caspart.CasPart:
		bytes, err := v.Write()
		if err != nil {
			s.panic(err.Error())
		}
		resource := new(dbpf.Resource)
		resource.FromBytes(bytes)
		return resource
	default:
		s.panic("value type not unparsable")
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

func (s *session) mergeExclude(filter *keys.Filter, pack *dbpf.Package) {
	f := s.excludes[pack]
	s.excludes[pack] = keys.MergeFilters(f, filter)
}

func (s *session) list(list interface{}) []interface{} {
	switch l := list.(type) {
	case nil:
		return nil
	case []interface{}:
		return l
	case *dbpf.Package:
		return s.listResources(l)
	case combined.Tunable:
		tunables := make([]interface{}, 0)
		for _, t := range l.Tunables {
			tunables = append(tunables, t)
		}
		return tunables
	default:
		s.panic("variable type does not contain a list")
		return nil
	}
}

func (s *session) listResources(pack *dbpf.Package) []interface{} {
	resources := pack.ListResources(s.includes[pack], s.excludes[pack], nil)
	for _, p := range s.merges[pack] {
		p.ListResources(s.includes[pack], s.excludes[pack], resources)
	}
	list := make([]interface{}, 0)
	for _, resource := range resources {
		list = append(list, resource)
	}
	return list
}
