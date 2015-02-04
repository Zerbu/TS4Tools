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

package keys

type Filter struct {
	Types, Groups []uint32
	Instances     []uint64
}

func (f *Filter) Include(k Keyer) bool {
	key := k.Key()

	if f.Types != nil {
		ok := false
		for _, t := range f.Types {
			if t == key.Type {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}

	if f.Groups != nil {
		ok := false
		for _, g := range f.Groups {
			if g == key.Group {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}

	if f.Instances != nil {
		ok := false
		for _, i := range f.Instances {
			if i == key.Instance {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}

	return true
}

func MergeFilters(a, b *Filter) *Filter {
	var types, groups []uint32
	var instances []uint64

	if a == nil {
		return b
	}

	if b == nil {
		return a
	}

	if a.Types == nil {
		types = b.Types
	} else if b.Types == nil {
		types = a.Types
	} else {
		types = merge32(a.Types, b.Types)
	}

	if a.Groups == nil {
		groups = b.Groups
	} else if b.Groups == nil {
		groups = a.Groups
	} else {
		groups = merge32(a.Groups, b.Groups)
	}

	if a.Instances == nil {
		instances = b.Instances
	} else if b.Instances == nil {
		instances = a.Instances
	} else {
		instances = merge64(a.Instances, b.Instances)
	}

	return &Filter{types, groups, instances}
}

func merge32(a, b []uint32) []uint32 {
	merge := make([]uint32, 0)
	for _, n := range a {
		merge = append(merge, n)
	}
	for _, n := range b {
		ok := true
		for _, e := range merge {
			if e == n {
				ok = false
			}
		}
		if ok {
			merge = append(merge, n)
		}
	}
	return merge
}

func merge64(a, b []uint64) []uint64 {
	merge := make([]uint64, 0)
	for _, n := range a {
		merge = append(merge, n)
	}
	for _, n := range b {
		ok := true
		for _, e := range merge {
			if e == n {
				ok = false
			}
		}
		if ok {
			merge = append(merge, n)
		}
	}
	return merge
}
