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

package hasher

import (
	"fmt"

	"github.com/Fogity/TS4Libs/hash"
	"gopkg.in/qml.v1"
)

const (
	numberFormatHex = "hex"
	numberFormatDec = "dec"

	formatHex32 = "0x%08X"
	formatHex64 = "0x%016X"
	formatDec32 = "%v"
	formatDec64 = "%v"
)

type Data struct {
	Fnv24, Fnv32, Fnv32High, Fnv64, Fnv64High string
	Text, Format                              string
}

func (d *Data) ChangeText(text string) {
	d.Text = text
	d.Calculate()
}

func (d *Data) ChangeFormat(format string) {
	d.Format = format
	d.Calculate()
}

func (d *Data) Calculate() {
	var f32, f64 string

	switch d.Format {
	case numberFormatHex:
		f32 = formatHex32
		f64 = formatHex64
	default:
		f32 = formatDec32
		f64 = formatDec64
	}

	ui32 := hash.Fnv24(d.Text)
	d.Fnv24 = fmt.Sprintf(f32, ui32)
	qml.Changed(d, &d.Fnv24)

	ui32 = hash.Fnv32(d.Text)
	d.Fnv32 = fmt.Sprintf(f32, ui32)
	qml.Changed(d, &d.Fnv32)

	ui32 = hash.Fnv32HighBit(d.Text)
	d.Fnv32High = fmt.Sprintf(f32, ui32)
	qml.Changed(d, &d.Fnv32High)

	ui64 := hash.Fnv64(d.Text)
	d.Fnv64 = fmt.Sprintf(f64, ui64)
	qml.Changed(d, &d.Fnv64)

	ui64 = hash.Fnv64HighBit(d.Text)
	d.Fnv64High = fmt.Sprintf(f64, ui64)
	qml.Changed(d, &d.Fnv64High)
}

func CreateWindow() error {
	engine := qml.NewEngine()

	hasher, err := engine.LoadFile("qrc:///qml/hasher/Window.qml")
	if err != nil {
		return err
	}

	context := engine.Context()

	d := new(Data)
	d.ChangeFormat(numberFormatHex)
	context.SetVar("app", d)

	window := hasher.CreateWindow(nil)
	window.Show()

	return nil
}
