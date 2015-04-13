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

package converter

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/qml.v1"
)

const (
	formatHex32 = "0x%08X"
	formatHex64 = "0x%016X"
	formatDec32 = "%v"
	formatDec64 = "%v"

	resultInputInvalid  = "Invalid input"
	resultInputTooLarge = "Number too large"
)

var (
	matchHex = regexp.MustCompile("^0x[0-9a-fA-F]+$")
	matchDec = regexp.MustCompile("^[0-9]+$")
)

type Data struct {
	Result     string
	AlignRight bool
}

func (d *Data) ChangeText(text string) {
	str := strings.TrimSpace(text)

	if str == "" {
		d.Update("", false)
		return
	}

	if matchHex.MatchString(str) {
		var ui32 uint32
		if _, err := fmt.Sscan(str, &ui32); err == nil {
			d.Update(fmt.Sprintf(formatDec32, ui32), true)
			return
		}

		var ui64 uint64
		if _, err := fmt.Sscan(str, &ui64); err == nil {
			d.Update(fmt.Sprintf(formatDec64, ui64), true)
			return
		}

		d.Update(resultInputTooLarge, false)
		return
	}

	if matchDec.MatchString(str) {
		str = strings.TrimLeft(str, "0")
		if str == "" {
			str = "0"
		}

		var ui32 uint32
		if _, err := fmt.Sscan(str, &ui32); err == nil {
			d.Update(fmt.Sprintf(formatHex32, ui32), true)
			return
		}

		var ui64 uint64
		if _, err := fmt.Sscan(str, &ui64); err == nil {
			d.Update(fmt.Sprintf(formatHex64, ui64), true)
			return
		}

		d.Update(resultInputTooLarge, false)
		return
	}

	d.Update(resultInputInvalid, false)
}

func (d *Data) Update(result string, alignRight bool) {
	d.Result = result
	d.AlignRight = alignRight
	qml.Changed(d, &d.Result)
}

func CreateWindow() error {
	engine := qml.NewEngine()

	converter, err := engine.LoadFile("qrc:///qml/converter/Window.qml")
	if err != nil {
		return err
	}

	context := engine.Context()
	context.SetVar("app", new(Data))

	window := converter.CreateWindow(nil)
	window.Show()

	return nil
}
