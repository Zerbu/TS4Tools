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

	"github.com/Fogity/TS4Tools/moddertoolbox/hasher"
	"gopkg.in/qml.v1"
)

type Dummy struct{}

func (d *Dummy) Create(tool string) {
	switch tool {
	case "hasher":
		hasher.CreateWindow()
	}
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

	toolbox, err := engine.LoadFile("qrc:///qml/Window.qml")
	if err != nil {
		return err
	}

	context := engine.Context()
	context.SetVar("dummy", new(Dummy))

	window := toolbox.CreateWindow(nil)
	window.Show()
	window.Wait()

	return nil
}
