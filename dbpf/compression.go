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

import "encoding/binary"

func internalDecompress(source []byte) ([]byte, error) {
	var data []byte
	var sp, dp int

	if (source[0] & 0x80) != 0 {
		size := binary.BigEndian.Uint32(source[2:6])
		data = make([]byte, size)
		sp = 6
	} else {
		size := binary.BigEndian.Uint32([]byte{0, source[2], source[3], source[4]})
		data = make([]byte, size)
		sp = 5
	}

	var end bool

	for !end {
		b0 := int(source[sp])
		sp++

		var sn, dn, do int

		switch {
		case b0 < 0x80:
			b1 := int(source[sp])
			sp++

			sn = b0 & 0x03
			dn = ((b0 & 0x1C) >> 2) + 3
			do = ((b0 & 0x60) << 3) + b1 + 1
		case b0 < 0xC0:
			b1 := int(source[sp])
			sp++
			b2 := int(source[sp])
			sp++

			sn = ((b1 & 0xC0) >> 6) & 0x03
			dn = (b0 & 0x3F) + 4
			do = ((b1 & 0x3F) << 8) + b2 + 1
		case b0 < 0xE0:
			b1 := int(source[sp])
			sp++
			b2 := int(source[sp])
			sp++
			b3 := int(source[sp])
			sp++

			sn = b0 & 0x03
			dn = ((b0 & 0x0C) << 6) + b3 + 5
			do = ((b0 & 0x10) << 12) + (b1 << 8) + b2 + 1
		case b0 < 0xFC:
			sn = ((b0 & 0x1F) << 2) + 4
			dn = 0
			do = 0
		default:
			sn = (b0 & 0x03)
			dn = 0
			do = 0
			end = true
		}

		for i := 0; i < sn; i++ {
			data[dp] = source[sp]
			dp++
			sp++
		}

		for i := 0; i < dn; i++ {
			data[dp] = data[dp-do]
			dp++
		}
	}

	return data, nil
}
