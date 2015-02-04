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
	"github.com/Fogity/TS4Tools/hash"
	"github.com/Fogity/TS4Tools/keys"
)

func parse(p *parser) action {
	actions := make([]action, 0)
	for p.hasMore() {
		a := parseAction(p)
		actions = append(actions, a)
	}
	return func(s *session) {
		for _, action := range actions {
			if action != nil {
				action(s)
			}
		}
	}
}

func parseAction(p *parser) action {
	word := p.word()
	switch word {
	case "":
		p.end()
		return nil

	case "for":
		temp := p.word()
		p.ensure("in")
		lister := parseExpression(p)
		actioner := parseExpression(p)
		return func(s *session) {
			fmt.Printf("starting for\n")
			list := s.list(lister(s))
			action := actioner(s).(action)
			for _, item := range list {
				session := newSession()
				session.parent = s
				session.vars[temp] = item
				action(session)
				session.close()
			}
			fmt.Printf("ending for\n")
		}

	case "set":
		name := p.name()
		p.ensure("to")
		valuer := parseExpression(p)
		return func(s *session) {
			fmt.Printf("setting %v\n", name)
			value := valuer(s)
			s.set(name, value)
		}

	case "open":
		pather := parseExpression(p)
		p.ensure("as")
		name := p.word()
		p.end()
		return func(s *session) {
			fmt.Printf("opening %v\n", name)
			path := pather(s).(string)
			pack, err := dbpf.Open(path)
			if err != nil {
				panic(err)
			}
			s.introduce(name, pack)
		}

	case "merge":
		pather := parseExpression(p)
		p.ensure("with")
		packer := parseExpression(p)
		p.end()
		return func(s *session) {
			fmt.Printf("merging\n")
			path := pather(s).(string)
			merge, err := dbpf.Open(path)
			if err != nil {
				panic(err)
			}
			pack := packer(s).(*dbpf.Package)
			s.mergePackage(pack, merge)
		}

	case "create":
		name := p.word()
		p.end()
		return func(s *session) {
			fmt.Printf("creating %v\n", name)
			pack := dbpf.New()
			s.introduce(name, pack)
		}

	case "save":
		packer := parseExpression(p)
		p.ensure("as")
		pather := parseExpression(p)
		p.end()
		return func(s *session) {
			fmt.Printf("saving\n")
			pack := packer(s).(*dbpf.Package)
			path := pather(s).(string)
			err := pack.SaveAs(path)
			if err != nil {
				s.panic(err.Error())
			}
		}

	case "include":
		filterer := parseExpression(p)
		p.ensure("from")
		packer := parseExpression(p)
		p.end()
		return func(s *session) {
			fmt.Printf("including\n")
			filter := filterer(s).(*keys.Filter)
			pack := packer(s).(*dbpf.Package)
			s.mergeInclude(filter, pack)
		}

	case "add":
		resourcer := parseExpression(p)
		p.ensure("to")
		packer := parseExpression(p)
		p.end()
		return func(s *session) {
			fmt.Printf("adding\n")
			resource := resourcer(s).(*dbpf.Resource)
			pack := packer(s).(*dbpf.Package)
			pack.AddResource(resource)
		}

	case "new":
		kind := p.word()
		p.ensure("as")
		name := p.word()
		p.end()
		return func(s *session) {
			fmt.Printf("new %v %v\n", kind, name)
			resource := s.newResource(kind)
			s.introduce(name, resource)
		}

	case "parse":
		resourcer := parseExpression(p)
		p.ensure("to")
		kind := p.word()
		p.ensure("as")
		name := p.word()
		p.end()
		return func(s *session) {
			fmt.Printf("parsing %v %v\n", kind, name)
			resource := resourcer(s).(*dbpf.Resource)
			parsed := s.parseResource(resource, kind)
			s.introduce(name, parsed)
		}

	default:
		p.panic("action '%v' not recognized", word)
		return nil
	}
}

func parseExpression(p *parser) expression {
	switch p.next() {
	case '"':
		str := p.string()
		return func(s *session) interface{} {
			return str
		}
	case '[':
		group := p.group('[', ']')
		con := parseConstruction(group)
		return func(s *session) interface{} {
			return con(s)
		}
	case '{':
		group := p.group('{', '}')
		action := parse(group)
		return func(s *session) interface{} {
			return action
		}
	default:
		name := p.name()
		return func(s *session) interface{} {
			return s.fetch(name)
		}
	}
}

func parseConstruction(p *parser) construction {
	word := p.word()
	switch word {
	case "":
		p.end()
		return nil

	case "group":
		hasher := parseExpression(p)
		p.end()
		return func(s *session) interface{} {
			switch h := hasher(s).(type) {
			case uint32:
				return &keys.Filter{nil, []uint32{h}, nil}
			case string:
				return &keys.Filter{nil, []uint32{hash.Fnv24(h)}, nil}
			default:
				p.panic("can not construct group hash of given input")
				return nil
			}
		}

	case "instance":
		hasher := parseExpression(p)
		p.end()
		return func(s *session) interface{} {
			switch h := hasher(s).(type) {
			case uint64:
				return &keys.Filter{nil, nil, []uint64{h}}
			case string:
				return &keys.Filter{nil, nil, []uint64{hash.Fnv64(h), hash.Fnv64HighBit(h)}}
			default:
				p.panic("can not construct instance hash of given input")
				return nil
			}
		}

	default:
		p.panic("construction '%v' not recognized", word)
		return nil
	}
}
