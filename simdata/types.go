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

package simdata

const (
	dtBool = iota
	dtChar8
	dtInt8
	dtUInt8
	dtInt16
	dtUInt16
	dtInt32
	dtUInt32
	dtInt64
	dtUInt64
	dtFloat
	dtString8
	dtHashedString8
	dtObject
	dtVector
	dtFloat2
	dtFloat3
	dtFloat4
	dtTableSetReference
	dtResourceKey
	dtLocKey
	dtUndefined
)
