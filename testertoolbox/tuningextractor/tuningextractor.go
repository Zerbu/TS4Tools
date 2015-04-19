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

package tuningextractor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Fogity/TS4Libs/caspart"
	"github.com/Fogity/TS4Libs/consts"
	"github.com/Fogity/TS4Libs/dbpf"
	"github.com/Fogity/TS4Libs/hash"
	"github.com/Fogity/TS4Libs/keys"
	"github.com/Fogity/TS4Libs/stbl"
	"github.com/Fogity/TS4Libs/tuning"
	"github.com/Fogity/TS4Libs/tuning/combined"
	"gopkg.in/qml.v1"
)

const (
	gameDirMissing   = "The Game Directory must be specified."
	exportDirMissing = "An Export Directory must be specified."
)

func trimPath(path string) string {
	return strings.TrimPrefix(path, "file:/")
}

func isPack(name string) bool {
	return strings.HasPrefix(name, "FP") || strings.HasPrefix(name, "GP") || strings.HasPrefix(name, "EP") || strings.HasPrefix(name, "SP")
}

func loadCombinedTunings(folder string) (map[string]*combined.Combined, map[int]string, error) {
	infos, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, nil, err
	}

	cts := make(map[string]*combined.Combined)
	tunings := make(map[int]string)
	for _, info := range infos {
		if path.Ext(info.Name()) != ".62e94d38" {
			continue
		}
		file, err := os.Open(fmt.Sprintf("%v/%v", folder, info.Name()))
		if err != nil {
			return nil, nil, err
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, nil, err
		}
		file.Close()
		ct, err := combined.Read(data)
		if err != nil {
			return nil, nil, err
		}
		group := info.Name()[:strings.Index(info.Name(), "!")]
		cts[group] = ct
		for _, e := range ct.Entries {
			for _, i := range e.Instances {
				var s int
				fmt.Sscan(i.Id, &s)
				if s != 0 {
					tunings[s] = i.Name
				}
			}
			for _, m := range e.Modules {
				var s int
				fmt.Sscan(m.Id, &s)
				if s != 0 {
					tunings[s] = m.Name
				}
			}
		}
	}

	return cts, tunings, nil
}

func loadStrings(folder string) (map[int]string, error) {
	infos, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	addons := make([]string, 0)
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		if isPack(info.Name()) {
			addons = append(addons, info.Name())
		}
	}

	strs := make(map[int]string)

	pack, err := dbpf.Open(fmt.Sprintf("%v/Data/Client/Strings_ENG_US.package", folder))
	if err != nil {
		return nil, err
	}
	for _, r := range pack.ListResources(nil, nil, nil) {
		data, err := r.ToBytes()
		if err != nil {
			return nil, err
		}
		table, err := stbl.Read(data)
		if err != nil {
			return nil, err
		}
		for _, e := range table.Entries {
			strs[int(e.Key)] = e.String
		}
	}

	for _, addon := range addons {
		pack, err := dbpf.Open(fmt.Sprintf("%v/Delta/%v/Strings_ENG_US.package", folder, addon))
		if err != nil {
			pack, err = dbpf.Open(fmt.Sprintf("%v/%v/Strings_ENG_US.package", folder, addon))
			if err != nil {
				continue
			}
		}
		for _, r := range pack.ListResources(nil, nil, nil) {
			data, err := r.ToBytes()
			if err != nil {
				return nil, err
			}
			table, err := stbl.Read(data)
			if err != nil {
				return nil, err
			}
			for _, e := range table.Entries {
				strs[int(e.Key)] = e.String
			}
		}
	}

	return strs, nil
}

func loadCasPartNames(folder string) (map[int]string, error) {
	infos, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	addons := make([]string, 0)
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		if isPack(info.Name()) {
			addons = append(addons, info.Name())
		}
	}

	filter := &keys.Filter{[]uint32{consts.ResourceTypeCasPart}, nil, nil}

	pack, err := dbpf.Open(fmt.Sprintf("%v/Data/Client/ClientFullBuild0.package", folder))
	if err != nil {
		return nil, err
	}
	list := pack.ListResources(filter, nil, nil)
	pack, err = dbpf.Open(fmt.Sprintf("%v/Data/Client/ClientDeltaBuild0.package", folder))
	if err != nil {
		return nil, err
	}
	list = pack.ListResources(filter, nil, list)

	for _, addon := range addons {
		pack, err = dbpf.Open(fmt.Sprintf("%v/%v/ClientFullBuild0.package", folder, addon))
		if err != nil {
			fmt.Println(err)
			continue
		}
		list = pack.ListResources(filter, nil, list)
		pack, err = dbpf.Open(fmt.Sprintf("%v/Delta/%v/ClientDeltaBuild0.package", folder, addon))
		if err != nil {
			fmt.Println(err)
			continue
		}
		list = pack.ListResources(filter, nil, list)
	}

	names := make(map[int]string)

	for k, r := range list {
		data, err := r.ToBytes()
		if err != nil {
			fmt.Println(err)
			continue
		}
		part, err := caspart.Read(data)
		if err != nil {
			fmt.Println(err)
			continue
		}
		names[int(k.Instance)] = part.Name
	}

	return names, nil
}

func formatName(instance combined.Instance, group uint32) string {
	var t uint32
	if instance.XMLName.Local == "M" {
		t = consts.ResourceTypeTuningModule
	} else {
		t = hash.Fnv32(instance.Instance)
	}
	var i uint64
	fmt.Sscan(instance.Id, &i)
	return fmt.Sprintf("S4_%08X_%08X_%016X", t, group, i)
}

type Data struct {
	GameDir, ExportDir, Information string
}

func (d *Data) inform(text string) {
	d.Information = text
	qml.Changed(d, &d.Information)
}

func (d *Data) report(err error) {
	d.Information = err.Error()
	qml.Changed(d, &d.Information)
}

func (d *Data) Export() {
	if d.GameDir == "" {
		d.inform(gameDirMissing)
		return
	}

	if d.ExportDir == "" {
		d.inform(exportDirMissing)
		return
	}

	gameFolder := trimPath(d.GameDir)
	exportFolder := trimPath(d.ExportDir)

	cts, tunings, err := loadCombinedTunings(exportFolder)
	if err != nil {
		d.report(err)
		return
	}

	strs, err := loadStrings(gameFolder)
	if err != nil {
		d.report(err)
		return
	}

	names, err := loadCasPartNames(gameFolder)
	if err != nil {
		d.report(err)
		return
	}

	context := new(tuning.Context)
	context.Indentation = "\t"
	context.LineEnd = "\n"
	context.AddReferences = true
	context.Strings = strs
	context.Tunings = tunings
	context.CasParts = names

	count := 0

	for group, ct := range cts {
		dir := fmt.Sprintf("%v/%v", exportFolder, group)
		var g uint32
		fmt.Sscan(group, &g)
		os.Mkdir(dir, 0700)
		for _, entry := range ct.Entries {
			dir := fmt.Sprintf("%v/%v", dir, entry.Type)
			os.Mkdir(dir, 0700)
			for _, inst := range entry.Instances {
				name := formatName(inst, g)
				path := fmt.Sprintf("%v/%v.xml", dir, name)
				file, err := os.Create(path)
				if err != nil {
					d.report(err)
					return
				}
				context.File = file
				err = context.Write(inst)
				if err != nil {
					d.report(err)
					return
				}
				file.Close()
				count++
			}
			for _, inst := range entry.Modules {
				name := formatName(inst, g)
				path := fmt.Sprintf("%v/%v.xml", dir, name)
				file, err := os.Create(path)
				if err != nil {
					d.report(err)
					return
				}
				context.File = file
				err = context.Write(inst)
				if err != nil {
					d.report(err)
					return
				}
				file.Close()
				count++
			}
		}
	}

	d.inform(fmt.Sprintf("Extraction completed, %v files extracted.", count))
}

func CreateWindow() error {
	engine := qml.NewEngine()

	extractor, err := engine.LoadFile("qrc:///qml/tuningextractor/Window.qml")
	if err != nil {
		return err
	}

	context := engine.Context()
	d := new(Data)
	d.Information = "Enter files and press Extract"
	context.SetVar("app", d)

	window := extractor.CreateWindow(nil)
	window.Show()

	return nil
}
