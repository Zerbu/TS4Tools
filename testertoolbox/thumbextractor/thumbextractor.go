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

package thumbextractor

import (
	"fmt"
	"os"
	"strings"

	"github.com/Fogity/TS4Libs/caspart"
	"github.com/Fogity/TS4Libs/consts"
	"github.com/Fogity/TS4Libs/dbpf"
	"github.com/Fogity/TS4Libs/keys"
	"github.com/Fogity/TS4Libs/thumbnail"
	"gopkg.in/qml.v1"
)

const (
	casPartFileMissing = "A Cas Part Package must be specified."
	thumbFileMissing   = "A Thumbnail Package must be specified."
	exportDirMissing   = "An Export Directory must be specified."
)

func trimPath(path string) string {
	return strings.TrimPrefix(path, "file:/")
}

type Data struct {
	CasPartFile, ThumbFile, ExportDir, Information string
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
	if d.CasPartFile == "" {
		d.inform(casPartFileMissing)
		return
	}

	if d.ThumbFile == "" {
		d.inform(thumbFileMissing)
		return
	}

	if d.ExportDir == "" {
		d.inform(exportDirMissing)
		return
	}

	casPartPack, err := dbpf.Open(trimPath(d.CasPartFile))
	if err != nil {
		d.report(err)
		return
	}

	thumbPack, err := dbpf.Open(trimPath(d.ThumbFile))
	if err != nil {
		d.report(err)
		return
	}

	folder := trimPath(d.ExportDir)

	casParts := make([]uint64, 0)
	casPartNames := make(map[uint64]string)
	for k, r := range casPartPack.ListResources(&keys.Filter{[]uint32{consts.ResourceTypeCasPart}, nil, nil}, nil, nil) {
		data, err := r.ToBytes()
		if err != nil {
			d.report(err)
			return
		}
		casPart, err := caspart.Read(data)
		if err != nil {
			d.report(err)
			return
		}
		casParts = append(casParts, k.Instance)
		casPartNames[k.Instance] = casPart.Name
	}

	count := 0
	for k, r := range thumbPack.ListResources(&keys.Filter{nil, []uint32{consts.ResourceGroupPortraitFemale, consts.ResourceGroupPortraitMale}, casParts}, nil, nil) {
		data, err := r.ToBytes()
		if err != nil {
			d.report(err)
			return
		}
		thumb, err := thumbnail.Convert(data)
		if err != nil {
			d.report(err)
			return
		}
		file, err := os.Create(fmt.Sprintf("%v/%v_%x.png", folder, casPartNames[k.Instance], k.Group))
		if err != nil {
			d.report(err)
			return
		}
		_, err = file.Write(thumb)
		if err != nil {
			d.report(err)
			file.Close()
			return
		}
		file.Close()
		count++
	}

	d.inform(fmt.Sprintf("Extraction completed, %v thumbnails extracted.", count))
}

func CreateWindow() error {
	engine := qml.NewEngine()

	extractor, err := engine.LoadFile("qrc:///qml/thumbextractor/Window.qml")
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
