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
	"fmt"

	"github.com/Fogity/GoLibs/script/context"
	"github.com/Fogity/GoLibs/script/define"
	"github.com/Fogity/GoLibs/script/parse"
	"github.com/Fogity/TS4Tools/caspart"
	"github.com/Fogity/TS4Tools/combined"
	"github.com/Fogity/TS4Tools/dbpf"
	"github.com/Fogity/TS4Tools/hash"
	"github.com/Fogity/TS4Tools/keys"
	"github.com/Fogity/TS4Tools/simdata"
)

var (
	definitions = &context.Definitions{}

	additional = context.Definitions{
		Values: map[string]interface{}{},
		Types: []*context.Type{
			pack, key, sdata, resource, casp, comb, combInst, combTune,
		},
		Statements: []*context.Statement{
			open, merge, create, save, include, exclude, parseRes, unparse, add,
		},
		Expressions: []*context.Expression{
			filter, search,
		},
	}

	pack = &context.Type{
		Name:  "package",
		Match: matchPack,
		List:  listPack,
	}

	key = &context.Type{
		Name: "key",
		Match: func(value interface{}) bool {
			_, ok := value.(keys.Key)
			return ok
		},
		Get: func(v interface{}, attr string) (interface{}, error) {
			k := v.(keys.Key)
			switch attr {
			case "type":
				return int64(k.Type), nil
			case "group":
				return int64(k.Group), nil
			case "instance":
				return int64(k.Instance), nil
			default:
				return nil, fmt.Errorf("type key does not contain value %v", attr)
			}
		},
	}

	resource = &context.Type{
		Name: "resource",
		Match: func(value interface{}) bool {
			_, ok := value.(*dbpf.Resource)
			return ok
		},
		Get: func(v interface{}, attr string) (interface{}, error) {
			res := v.(*dbpf.Resource)
			switch attr {
			case "key":
				return res.Key(), nil
			default:
				return nil, fmt.Errorf("resource does not contain value %v", attr)
			}
		},
		Set: func(v interface{}, attr string, value interface{}) error {
			res := v.(*dbpf.Resource)
			switch attr {
			case "key":
				key, ok := value.(keys.Key)
				if !ok {
					return fmt.Errorf("expected key value")
				}
				res.SetKey(key)
				return nil
			default:
				return fmt.Errorf("resource does not contain value %v", attr)
			}
		},
	}

	sdata = &context.Type{
		Name: "simdata",
		Match: func(value interface{}) bool {
			_, ok := value.(*simdata.Simdata)
			return ok
		},
		Get: func(v interface{}, attr string) (interface{}, error) {
			data := v.(*simdata.Simdata)
			value, ok := data.GetValue(attr)
			if !ok {
				return nil, fmt.Errorf("simdata does not contain value %v", attr)
			}
			return value, nil
		},
		Set: func(v interface{}, attr string, value interface{}) error {
			data := v.(*simdata.Simdata)
			return data.SetValue(attr, value)
		},
	}

	casp = &context.Type{
		Name: "caspart",
		Match: func(value interface{}) bool {
			_, ok := value.(*caspart.CasPart)
			return ok
		},
		Get: func(v interface{}, attr string) (interface{}, error) {
			data := v.(*caspart.CasPart)
			switch attr {
			case "Name":
				return data.Name, nil
			case "ShowInUI":
				return (data.ParamFlags & caspart.ShowInUI) != 0, nil
			default:
				return nil, fmt.Errorf("caspart does not have the attribute %v", attr)
			}
		},
		Set: func(v interface{}, attr string, value interface{}) error {
			data := v.(*caspart.CasPart)
			switch attr {
			case "Name":
				val, ok := value.(string)
				if !ok {
					return fmt.Errorf("expected string value")
				}
				data.Name = val
				return nil
			case "ShowInUI":
				val, ok := value.(bool)
				if !ok {
					return fmt.Errorf("expected boolean value")
				}
				if val {
					data.ParamFlags |= caspart.ShowInUI
					return nil
				}
				data.ParamFlags &= ^uint8(caspart.ShowInUI)
				return nil
			default:
				return fmt.Errorf("caspart does not have the attribute %v", attr)
			}
		},
	}

	comb = &context.Type{
		Name: "combined",
		Match: func(value interface{}) bool {
			_, ok := value.(*combined.Combined)
			return ok
		},
		Get: func(v interface{}, attr string) (interface{}, error) {
			data := v.(*combined.Combined)
			for _, entry := range data.Entries {
				if entry.Type == attr {
					list := make([]interface{}, len(entry.Instances))
					for i, item := range entry.Instances {
						list[i] = item
					}
					return list, nil
				}
			}
			return nil, fmt.Errorf("no entry of type %v", attr)
		},
	}

	combInst = &context.Type{
		Name: "instance",
		Match: func(value interface{}) bool {
			_, ok := value.(combined.Instance)
			return ok
		},
		Get: func(v interface{}, attr string) (interface{}, error) {
			data := v.(combined.Instance)
			for _, tunable := range data.Tunables {
				if tunable.Name == attr {
					return tunable, nil
				}
			}
			return nil, fmt.Errorf("tunable named %v not found", attr)
		},
	}

	combTune = &context.Type{
		Name: "tunable",
		Match: func(value interface{}) bool {
			_, ok := value.(combined.Tunable)
			return ok
		},
		Get: func(v interface{}, attr string) (interface{}, error) {
			data := v.(combined.Tunable)
			switch attr {
			case "Value":
				return data.Value, nil
			default:
				return nil, fmt.Errorf("type tunable does not have attribute %v", attr)
			}
		},
		List: func(value interface{}) ([]interface{}, error) {
			data := value.(combined.Tunable)
			list := make([]interface{}, len(data.Tunables))
			for i, item := range data.Tunables {
				list[i] = item
			}
			return list, nil
		},
	}

	open = &context.Statement{
		Name:      "open",
		Match:     matchOpen,
		Arguments: argsOpen,
		Execute:   execOpen,
	}

	merge = &context.Statement{
		Name:      "merge",
		Match:     matchMerge,
		Arguments: argsMerge,
		Execute:   execMerge,
	}

	include = &context.Statement{
		Name:      "include",
		Match:     matchInc,
		Arguments: argsInc,
		Execute:   execInc,
	}

	exclude = &context.Statement{
		Name:      "exclude",
		Match:     matchExc,
		Arguments: argsExc,
		Execute:   execExc,
	}

	create = &context.Statement{
		Name:      "create",
		Match:     matchCreate,
		Arguments: argsCreate,
		Execute:   execCreate,
	}

	save = &context.Statement{
		Name:      "save",
		Match:     matchSave,
		Arguments: argsSave,
		Execute:   execSave,
	}

	parseRes = &context.Statement{
		Name:      "parse",
		Match:     matchParse,
		Arguments: argsParse,
		Execute:   execParse,
	}

	unparse = &context.Statement{
		Name:      "unparse",
		Match:     matchUnparse,
		Arguments: argsUnparse,
		Execute:   execUnparse,
	}

	add = &context.Statement{
		Name:      "add",
		Match:     matchAdd,
		Arguments: argsAdd,
		Execute:   execAdd,
	}

	filter = &context.Expression{
		Name:      "filter",
		Match:     matchFilter,
		Arguments: argsFilter,
		Execute:   execFilter,
	}

	search = &context.Expression{
		Name:      "search",
		Match:     matchSearch,
		Arguments: argsSearch,
		Execute:   execSearch,
	}
)

func init() {
	types := make([]*context.Type, 0)
	stmts := make([]*context.Statement, 0)
	exprs := make([]*context.Expression, 0)
	vals := make(map[string]interface{})

	for _, t := range define.Standard.Types {
		types = append(types, t)
	}
	for _, s := range define.Standard.Statements {
		stmts = append(stmts, s)
	}
	for _, e := range define.Standard.Expressions {
		exprs = append(exprs, e)
	}
	for k, v := range define.Standard.Values {
		vals[k] = v
	}

	for _, t := range additional.Types {
		types = append(types, t)
	}
	for _, s := range additional.Statements {
		stmts = append(stmts, s)
	}
	for _, e := range additional.Expressions {
		exprs = append(exprs, e)
	}
	for k, v := range additional.Values {
		vals[k] = v
	}

	definitions.Types = types
	definitions.Statements = stmts
	definitions.Expressions = exprs
	definitions.Values = vals
}

func matchPack(value interface{}) bool {
	_, ok := value.(*Package)
	return ok
}

func listPack(value interface{}) ([]interface{}, error) {
	v := value.(*Package)
	resources := v.p.ListResources(v.include, v.exclude, nil)
	for _, p := range v.merge {
		p.ListResources(v.include, v.exclude, resources)
	}
	list := make([]interface{}, 0)
	for _, resource := range resources {
		list = append(list, resource)
	}
	return list, nil
}

func matchOpen(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "open") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "as") && define.IsName(parts[3])
}

func argsOpen(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execOpen(c *context.Context, args []interface{}) error {
	pather, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	path, ok := pather.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	name := define.GetName(args[1])

	p, err := OpenPackage(path)
	if err != nil {
		return err
	}

	c.Set(name, p)

	return nil
}

func matchMerge(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "merge") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "with") && define.IsExpression(parts[3])
}

func argsMerge(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execMerge(c *context.Context, args []interface{}) error {
	pather, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	path, ok := pather.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	packer, err := define.GetArgument(c, args[1])
	if err != nil {
		return err
	}

	p, ok := packer.(*Package)
	if !ok {
		return fmt.Errorf("expected package value")
	}

	return p.Merge(path)
}

func matchInc(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "include") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "from") && define.IsExpression(parts[3])
}

func argsInc(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execInc(c *context.Context, args []interface{}) error {
	filterer, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	filter, ok := filterer.(*keys.Filter)
	if !ok {
		return fmt.Errorf("expected filter value")
	}

	packer, err := define.GetArgument(c, args[1])
	if err != nil {
		return err
	}

	p, ok := packer.(*Package)
	if !ok {
		return fmt.Errorf("expected package value")
	}

	p.Include(filter)

	return nil
}

func matchExc(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "exclude") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "from") && define.IsExpression(parts[3])
}

func argsExc(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execExc(c *context.Context, args []interface{}) error {
	filterer, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	filter, ok := filterer.(*keys.Filter)
	if !ok {
		return fmt.Errorf("expected filter value")
	}

	packer, err := define.GetArgument(c, args[1])
	if err != nil {
		return err
	}

	p, ok := packer.(*Package)
	if !ok {
		return fmt.Errorf("expected package value")
	}

	p.Exclude(filter)

	return nil
}

func matchCreate(parts []parse.Part) bool {
	if len(parts) != 2 {
		return false
	}

	return define.IsKeyword(parts[0], "create") && define.IsName(parts[1])
}

func argsCreate(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1]}
}

func execCreate(c *context.Context, args []interface{}) error {
	name := define.GetName(args[0])

	c.Set(name, CreatePackage())

	return nil
}

func matchParse(parts []parse.Part) bool {
	if len(parts) != 6 {
		return false
	}

	return define.IsKeyword(parts[0], "parse") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "to") && define.IsName(parts[3]) && define.IsKeyword(parts[4], "as") && define.IsName(parts[5])
}

func argsParse(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3], parts[5]}
}

func execParse(c *context.Context, args []interface{}) error {
	reser, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	res, ok := reser.(*dbpf.Resource)
	if !ok {
		return fmt.Errorf("expected resource value")
	}

	kind := define.GetName(args[1])

	name := define.GetName(args[2])

	switch kind {
	case "simdata":
		bytes, err := res.ToBytes()
		if err != nil {
			return err
		}
		data, err := simdata.Read(bytes)
		if err != nil {
			return err
		}
		c.Set(name, data)
		return nil
	case "caspart":
		bytes, err := res.ToBytes()
		if err != nil {
			return err
		}
		data, err := caspart.Read(bytes)
		if err != nil {
			return err
		}
		c.Set(name, data)
		return nil
	case "combined":
		bytes, err := res.ToBytes()
		if err != nil {
			return err
		}
		data, err := combined.Read(bytes)
		if err != nil {
			return err
		}
		c.Set(name, data)
		return nil
	default:
		return fmt.Errorf("unknown resource type")
	}
}

func matchUnparse(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "unparse") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "as") && define.IsName(parts[3])
}

func argsUnparse(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execUnparse(c *context.Context, args []interface{}) error {
	dater, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	name := define.GetName(args[1])

	switch data := dater.(type) {
	case *simdata.Simdata:
		bytes, err := data.Write()
		if err != nil {
			return err
		}
		res := new(dbpf.Resource)
		res.FromBytes(bytes)
		c.Set(name, res)
		return nil
	case *caspart.CasPart:
		bytes, err := data.Write()
		if err != nil {
			return err
		}
		res := new(dbpf.Resource)
		res.FromBytes(bytes)
		c.Set(name, res)
		return nil
	default:
		return fmt.Errorf("unknown resource type")
	}
}

func matchSave(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "save") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "as") && define.IsExpression(parts[3])
}

func argsSave(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execSave(c *context.Context, args []interface{}) error {
	packer, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	pack, ok := packer.(*Package)
	if !ok {
		return fmt.Errorf("expected package value")
	}

	pather, err := define.GetArgument(c, args[1])
	if err != nil {
		return err
	}

	path, ok := pather.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	return pack.p.SaveAs(path)
}

func matchAdd(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "add") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "to") && define.IsExpression(parts[3])
}

func argsAdd(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execAdd(c *context.Context, args []interface{}) error {
	reser, err := define.GetArgument(c, args[0])
	if err != nil {
		return err
	}

	res, ok := reser.(*dbpf.Resource)
	if !ok {
		return fmt.Errorf("expected resource value")
	}

	packer, err := define.GetArgument(c, args[1])
	if err != nil {
		return err
	}

	pack, ok := packer.(*Package)
	if !ok {
		return fmt.Errorf("expected package value")
	}

	pack.p.AddResource(res)

	return nil
}

func matchFilter(parts []parse.Part) bool {
	if len(parts) != 3 {
		return false
	}

	return define.IsKeyword(parts[0], "filter") && define.IsName(parts[1]) && define.IsExpression(parts[2])
}

func argsFilter(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[2]}
}

func execFilter(c *context.Context, args []interface{}) (interface{}, error) {
	kind := define.GetName(args[0])

	hasher, err := define.GetArgument(c, args[1])
	if err != nil {
		return nil, err
	}

	number, n := hasher.(int64)
	str, s := hasher.(string)

	if !n && !s {
		return nil, fmt.Errorf("expected string or integer")
	}

	switch kind {
	case "type":
		num := uint32(number)
		if !n {
			num = hash.Fnv32(str)
		}
		return &keys.Filter{[]uint32{num}, nil, nil}, nil
	case "group":
		num := uint32(number)
		if !n {
			num = hash.Fnv24(str)
		}
		return &keys.Filter{nil, []uint32{num}, nil}, nil
	case "instance":
		num := uint64(number)
		if !n {
			num = hash.Fnv64(str)
		}
		return &keys.Filter{nil, nil, []uint64{num}}, nil
	default:
		return nil, fmt.Errorf("unknown filter type")
	}
}

func matchSearch(parts []parse.Part) bool {
	if len(parts) != 4 {
		return false
	}

	return define.IsKeyword(parts[0], "search") && define.IsExpression(parts[1]) && define.IsKeyword(parts[2], "for") && define.IsExpression(parts[3])
}

func argsSearch(parts []parse.Part) []parse.Part {
	return []parse.Part{parts[1], parts[3]}
}

func execSearch(c *context.Context, args []interface{}) (interface{}, error) {
	tuner, err := define.GetArgument(c, args[0])
	if err != nil {
		return nil, err
	}

	tun, ok := tuner.(combined.Tunable)
	if !ok {
		return nil, fmt.Errorf("expected tunable value")
	}

	queryer, err := define.GetArgument(c, args[1])
	if err != nil {
		return nil, err
	}

	query, ok := queryer.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected search query")
	}

	return searchTunable(tun, query), nil
}

func searchTunable(tun combined.Tunable, query map[string]interface{}) []interface{} {
	list := make([]interface{}, 0)
	for _, t := range tun.Tunables {
		for _, item := range searchTunable(t, query) {
			list = append(list, item)
		}
	}
	match := true
	for k, v := range query {
		switch k {
		case "Tag":
			if tun.XMLName.Local != v {
				match = false
			}
		case "Name":
			if tun.Name != v {
				match = false
			}
		default:
			panic("unknown query")
		}
	}
	if match {
		list = append(list, tun)
	}
	return list
}
