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

package hash

import (
	"hash/fnv"
	"strings"
)

const (
	mask24    = (1 << 24) - 1
	highBit32 = 1 << (32 - 1)
	highBit64 = 1 << (64 - 1)
)

func Fnv24(s string) uint32 {
	hash := Fnv32(s)
	return (hash >> 24) ^ (hash & mask24)
}

func Fnv32(s string) uint32 {
	hash := fnv.New32()
	hash.Write([]byte(strings.ToLower(s)))
	return hash.Sum32()
}

func Fnv32HighBit(s string) uint32 {
	return Fnv32(s) | highBit32
}

func Fnv64(s string) uint64 {
	hash := fnv.New64()
	hash.Write([]byte(strings.ToLower(s)))
	return hash.Sum64()
}

func Fnv64HighBit(s string) uint64 {
	return Fnv64(s) | highBit64
}
