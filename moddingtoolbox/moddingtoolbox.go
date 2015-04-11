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

package main

//go:generate genqrc qml

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Fogity/TS4Libs/hash"
	"gopkg.in/qml.v1"
)

const (
	formatHex = "hex"
	formatDec = "dec"

	formatHex32 = "0x%08X"
	formatHex64 = "0x%016X"
	formatDec32 = "%v"
	formatDec64 = "%v"

	invalidInput = "Invalid input"
	largeInput   = "Number too large"
)

var (
	matchHex = regexp.MustCompile("^0x[a-zA-Z0-9]+$")
	matchDec = regexp.MustCompile("^[0-9]+$")
)

type Hash struct {
	Fnv24, Fnv32, Fnv32High, Fnv64, Fnv64High string
	Text, Format                              string
}

func (h *Hash) ChangeText(text string) {
	h.Text = text
	h.Calculate()
}

func (h *Hash) ChangeFormat(format string) {
	h.Format = format
	h.Calculate()
}

func (h *Hash) Calculate() {
	var f32, f64 string

	switch h.Format {
	case formatHex:
		f32 = formatHex32
		f64 = formatHex64
	default:
		f32 = formatDec32
		f64 = formatDec64
	}

	ui32 := hash.Fnv24(h.Text)
	h.Fnv24 = fmt.Sprintf(f32, ui32)
	qml.Changed(h, &h.Fnv24)

	ui32 = hash.Fnv32(h.Text)
	h.Fnv32 = fmt.Sprintf(f32, ui32)
	qml.Changed(h, &h.Fnv32)

	ui32 = hash.Fnv32HighBit(h.Text)
	h.Fnv32High = fmt.Sprintf(f32, ui32)
	qml.Changed(h, &h.Fnv32High)

	ui64 := hash.Fnv64(h.Text)
	h.Fnv64 = fmt.Sprintf(f64, ui64)
	qml.Changed(h, &h.Fnv64)

	ui64 = hash.Fnv64HighBit(h.Text)
	h.Fnv64High = fmt.Sprintf(f64, ui64)
	qml.Changed(h, &h.Fnv64High)
}

type Convert struct {
	Result     string
	AlignRight bool
}

func (c *Convert) ChangeText(text string) {
	str := strings.TrimSpace(text)

	if str == "" {
		c.Result = ""
		qml.Changed(c, &c.Result)
		return
	}

	if matchHex.MatchString(str) {
		var ui32 uint32
		if _, err := fmt.Sscan(str, &ui32); err == nil {
			c.Result = fmt.Sprintf(formatDec32, ui32)
			c.AlignRight = true
			qml.Changed(c, &c.Result)
			return
		}

		var ui64 uint64
		if _, err := fmt.Sscan(str, &ui64); err == nil {
			c.Result = fmt.Sprintf(formatDec64, ui64)
			c.AlignRight = true
			qml.Changed(c, &c.Result)
			return
		}

		c.Result = largeInput
		c.AlignRight = false
		qml.Changed(c, &c.Result)
		return
	}

	if matchDec.MatchString(str) {
		str = strings.TrimLeft(str, "0")
		if str == "" {
			str = "0"
		}

		var ui32 uint32
		if _, err := fmt.Sscan(str, &ui32); err == nil {
			c.Result = fmt.Sprintf(formatHex32, ui32)
			c.AlignRight = true
			qml.Changed(c, &c.Result)
			return
		}

		var ui64 uint64
		if _, err := fmt.Sscan(str, &ui64); err == nil {
			c.Result = fmt.Sprintf(formatHex64, ui64)
			c.AlignRight = true
			qml.Changed(c, &c.Result)
			return
		}

		c.Result = largeInput
		c.AlignRight = false
		qml.Changed(c, &c.Result)
		return
	}

	c.Result = invalidInput
	c.AlignRight = false
	qml.Changed(c, &c.Result)
}

func main() {
	if err := qml.Run(run); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	engine := qml.NewEngine()

	engine.On("quit", func() { os.Exit(0) })

	toolbox, err := engine.LoadFile("qrc:///qml/Toolbox.qml")
	if err != nil {
		return err
	}

	context := engine.Context()

	h := new(Hash)
	h.ChangeFormat(formatHex)
	context.SetVar("hash", h)

	context.SetVar("convert", &Convert{})

	window := toolbox.CreateWindow(nil)
	window.Show()
	window.Wait()

	return nil
}
