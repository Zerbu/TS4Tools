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

package caspart

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	version = 30
)

const (
	DefaultForBodyType   = 1 << 0
	DefaultThumbnailPart = 1 << 1
	AllowForRandom       = 1 << 2
	ShowInUI             = 1 << 3
	ShowInSimInfoPanel   = 1 << 4
	ShowInCasDemo        = 1 << 5
)

func Read(b []byte) (*CasPart, error) {
	r := bytes.NewReader(b)
	p := new(CasPart)

	err := binary.Read(r, binary.LittleEndian, &p.Chunk1)
	if err != nil {
		return nil, err
	}

	p.Presets = make([]Preset, p.NumPresets)
	for i := range p.Presets {
		var preset Preset
		err = binary.Read(r, binary.LittleEndian, &preset)
		if int(preset.NumParams) != 0 {
			return nil, fmt.Errorf("preset parameters not supported")
		}
		p.Presets[i] = preset
	}

	low, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	var size int
	if low > 127 {
		high, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		size = int((high << 7) | (low & 0x7F))
	} else {
		size = int(low)
	}
	bts := make([]byte, size/2)
	for i := range bts {
		r.ReadByte()
		t, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		bts[i] = t
	}
	p.Name = string(bts)

	err = binary.Read(r, binary.LittleEndian, &p.Chunk2)
	if err != nil {
		return nil, err
	}

	p.Tags = make([]Tag, p.NumTags)
	err = binary.Read(r, binary.LittleEndian, p.Tags)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &p.Chunk3)
	if err != nil {
		return nil, err
	}

	if p.Unused2 > 0 {
		err = binary.Read(r, binary.LittleEndian, &p.Unused3)
		if err != nil {
			return nil, err
		}
	}

	err = binary.Read(r, binary.LittleEndian, &p.NumSwatchColors)
	if err != nil {
		return nil, err
	}

	p.SwatchColors = make([]uint32, p.NumSwatchColors)
	err = binary.Read(r, binary.LittleEndian, p.SwatchColors)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &p.Chunk4)
	if err != nil {
		return nil, err
	}

	p.AuralMaterialSets = make([]uint32, p.UsedMaterialCount)
	err = binary.Read(r, binary.LittleEndian, p.AuralMaterialSets)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &p.Chunk5)
	if err != nil {
		return nil, err
	}

	p.LODs = make([]LOD, p.NumLODs)
	for i := range p.LODs {
		lod, err := readLOD(r)
		if err != nil {
			return nil, err
		}
		p.LODs[i] = *lod
	}

	err = binary.Read(r, binary.LittleEndian, &p.NumSlotKeys)
	if err != nil {
		return nil, err
	}

	p.SlotKeys = make([]uint8, p.NumSlotKeys)
	err = binary.Read(r, binary.LittleEndian, p.SlotKeys)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &p.Chunk6)
	if err != nil {
		return nil, err
	}

	p.RegionLayerOverrides = make([]Override, p.NumOverrides)
	err = binary.Read(r, binary.LittleEndian, p.RegionLayerOverrides)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &p.Chunk7)
	if err != nil {
		return nil, err
	}

	p.ResourceKeys = make([]ResourceKey, p.NumResourceKeys)
	err = binary.Read(r, binary.LittleEndian, p.ResourceKeys)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func readLOD(r *bytes.Reader) (*LOD, error) {
	var lod LOD
	err := binary.Read(r, binary.LittleEndian, &lod.Level)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &lod.Unused)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &lod.NumAssets)
	if err != nil {
		return nil, err
	}
	lod.Assets = make([]LODAsset, lod.NumAssets)
	err = binary.Read(r, binary.LittleEndian, lod.Assets)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &lod.NumLODKeys)
	if err != nil {
		return nil, err
	}
	lod.LODKeys = make([]uint8, lod.NumLODKeys)
	err = binary.Read(r, binary.LittleEndian, lod.LODKeys)
	if err != nil {
		return nil, err
	}
	return &lod, nil
}

func (p *CasPart) Write() ([]byte, error) {
	b := new(bytes.Buffer)

	err := binary.Write(b, binary.LittleEndian, p.Chunk1)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.Presets)
	if err != nil {
		return nil, err
	}

	size := len(p.Name) * 2
	if size > 127 {
		err = b.WriteByte(byte((size & 0x7F) | 0x80))
		if err != nil {
			return nil, err
		}
		err = b.WriteByte(byte(size >> 7))
		if err != nil {
			return nil, err
		}
	} else {
		err = b.WriteByte(byte(size))
		if err != nil {
			return nil, err
		}
	}
	for _, bt := range p.Name {
		err = b.WriteByte(0)
		if err != nil {
			return nil, err
		}
		err = b.WriteByte(byte(bt))
		if err != nil {
			return nil, err
		}
	}

	err = binary.Write(b, binary.LittleEndian, p.Chunk2)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.Tags)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.Chunk3)
	if err != nil {
		return nil, err
	}

	if p.Unused2 > 0 {
		err = binary.Write(b, binary.LittleEndian, p.Unused3)
		if err != nil {
			return nil, err
		}
	}

	err = binary.Write(b, binary.LittleEndian, p.NumSwatchColors)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.SwatchColors)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.Chunk4)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.AuralMaterialSets)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.Chunk5)
	if err != nil {
		return nil, err
	}

	for _, lod := range p.LODs {
		err = writeLOD(b, lod)
		if err != nil {
			return nil, err
		}
	}

	err = binary.Write(b, binary.LittleEndian, p.NumSlotKeys)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.SlotKeys)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.Chunk6)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.RegionLayerOverrides)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.Chunk7)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.LittleEndian, p.ResourceKeys)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func writeLOD(b *bytes.Buffer, lod LOD) error {
	err := binary.Write(b, binary.LittleEndian, lod.Level)
	if err != nil {
		return err
	}
	err = binary.Write(b, binary.LittleEndian, lod.Unused)
	if err != nil {
		return err
	}
	err = binary.Write(b, binary.LittleEndian, lod.NumAssets)
	if err != nil {
		return err
	}
	err = binary.Write(b, binary.LittleEndian, lod.Assets)
	if err != nil {
		return err
	}
	err = binary.Write(b, binary.LittleEndian, lod.NumLODKeys)
	if err != nil {
		return err
	}
	err = binary.Write(b, binary.LittleEndian, lod.LODKeys)
	if err != nil {
		return err
	}
	return nil
}

type CasPart struct {
	Chunk1
	Presets []Preset
	Name    string
	Chunk2
	Tags []Tag
	Chunk3
	Unused3         uint8
	NumSwatchColors uint8
	SwatchColors    []uint32
	Chunk4
	AuralMaterialSets []uint32
	Chunk5
	LODs        []LOD
	NumSlotKeys uint8
	SlotKeys    []uint8
	Chunk6
	RegionLayerOverrides []Override
	Chunk7
	ResourceKeys []ResourceKey
}

type Chunk1 struct {
	Version    uint32
	DataSize   uint32
	NumPresets uint32
}

type Chunk2 struct {
	DisplayIndex               float32
	SecondaryDislpayIndex      uint16
	PrototypeId                uint32
	AuralMaterialHash          uint32
	ParamFlags                 uint8
	ExcludePartFlags           uint64
	ExcludeModifierRegionFlags uint32
	NumTags                    uint32
}

type Chunk3 struct {
	SimoleonPrice      uint32
	PartTitleKey       uint32
	PartDescKey        uint32
	UniqueTextureSpace uint8
	BodyType           int32
	Unused1            int32
	AgeGender          uint32
	Unused2            uint8
}

type Chunk4 struct {
	BuffResKey        uint8
	VariantThumbKey   uint8
	VoiceEffectHash   uint64
	UsedMaterialCount uint8
}

type Chunk5 struct {
	NakedKey  uint8
	ParentKey uint8
	SortLayer int32
	NumLODs   uint8
}

type Chunk6 struct {
	DiffuseKey        uint8
	ShadowKey         uint8
	CompositionMethod uint8
	RegionMapKey      uint8
	NumOverrides      uint8
}

type Chunk7 struct {
	NormalMapKey     uint8
	SpecularMapKey   uint8
	NormalUVBodyType uint32
	EmissionMapKey   uint8
	NumResourceKeys  uint8
}

type Preset struct {
	CompleteId uint64
	NumParams  uint8
}

type Tag struct {
	Category, Value uint16
}

type LOD struct {
	Level      uint8
	Unused     uint32
	NumAssets  uint8
	Assets     []LODAsset
	NumLODKeys uint8
	LODKeys    []uint8
}

type LODAsset struct {
	Sorting, SpecLevel, CastShadow int32
}

type Override struct {
	Region uint8
	Layer  float32
}

type ResourceKey struct {
	Instance uint64
	Group    uint32
	Type     uint32
}
