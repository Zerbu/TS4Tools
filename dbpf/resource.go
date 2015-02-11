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

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/Fogity/TS4Tools/keys"
	"io/ioutil"
	"os"
)

const (
	compUncompressed = uint16(0x0000)
	compStreamable   = uint16(0xFFFE)
	compInternal     = uint16(0xFFFF)
	compDeleted      = uint16(0xFFE0)
	compZLIB         = uint16(0x5A42)
)

const (
	committed = 1
)

type Resource struct {
	key   keys.Key
	p     *Package
	entry *entry
	bytes []byte
	comp  []byte
}

func (r *Resource) Key() keys.Key {
	return r.key
}

func (r *Resource) SetKey(key keys.Key) {
	r.key = key
}

func (r *Resource) ToBytes() ([]byte, error) {
	switch r.entry.Extended.CompressionType {
	case compDeleted:
		return nil, nil
	case compZLIB:
		_, err := r.p.file.Seek(int64(r.entry.Fixed.Position), os.SEEK_SET)
		if err != nil {
			return nil, err
		}
		compressed := make([]byte, r.entry.Fixed.CompressedSize & ^uint32(extendedCompressionType))
		_, err = r.p.file.Read(compressed)
		if err != nil {
			return nil, err
		}
		reader, err := zlib.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, err
		}
		decompressed, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return decompressed, nil
	default:
		return nil, fmt.Errorf("unknown compression type %v", r.entry.Extended.CompressionType)
	}
}

func (r *Resource) FromBytes(b []byte) {
	r.bytes = b
}

func (p *Package) loadResourceList() {
	p.resources = make([]*Resource, p.header.EntryCount)

	isConstType := (p.record.Flags & constantType) != 0
	isConstGroup := (p.record.Flags & constantGroup) != 0
	isConstInstEx := (p.record.Flags & constantInstanceEx) != 0
	constType := p.record.Type
	constGroup := p.record.Group
	constInstEx := p.record.InstanceEx

	for i, entry := range p.record.Entries {
		var resource Resource

		var t, g, ie uint32
		if isConstType {
			t = constType
		} else {
			t = entry.Type
		}
		if isConstGroup {
			g = constGroup
		} else {
			g = entry.Group
		}
		if isConstInstEx {
			ie = constInstEx
		} else {
			ie = entry.InstanceEx
		}
		resource.key = keys.CombineKey(t, g, ie, entry.Fixed.Instance)
		resource.p = p
		resource.entry = entry

		p.resources[i] = &resource
	}
}

func (p *Package) saveResourceList() error {
	p.record.Entries = make([]*entry, len(p.resources))

	if len(p.record.Entries) == 0 {
		return nil
	}

	offset := uint32(headerSize)

	isConstType := true
	isConstGroup := true
	isConstInstEx := true
	constType := p.resources[0].key.Type
	constGroup := p.resources[0].key.Group
	constInstEx := uint32(p.resources[0].key.Instance >> 32)

	for i, resource := range p.resources {
		var entry entry

		key := resource.key
		t := key.Type
		g := key.Group
		ie := uint32(key.Instance >> 32)
		in := uint32(key.Instance)

		if t != constType {
			isConstType = false
		}
		if g != constGroup {
			isConstGroup = false
		}
		if ie != constInstEx {
			isConstInstEx = false
		}

		entry.Type = t
		entry.Group = g
		entry.InstanceEx = ie
		entry.Fixed.Instance = in

		if resource.bytes == nil {
			entry.Fixed.CompressedSize = extendedCompressionType
		} else {
			entry.Fixed.DecompressedSize = uint32(len(resource.bytes))

			buf := new(bytes.Buffer)
			w := zlib.NewWriter(buf)
			_, err := w.Write(resource.bytes)
			if err != nil {
				return err
			}
			w.Close()
			resource.comp = buf.Bytes()

			entry.Fixed.CompressedSize = uint32(len(resource.comp)) | extendedCompressionType

			entry.Extended.CompressionType = compZLIB

			entry.Fixed.Position = offset

			offset += uint32(len(resource.comp))
		}

		entry.Extended.Committed = committed

		p.record.Entries[i] = &entry
	}

	flags := uint32(0)
	num := 0
	if isConstType {
		flags = flags | constantType
		num++
	}
	if isConstGroup {
		flags = flags | constantGroup
		num++
	}
	if isConstInstEx {
		flags = flags | constantInstanceEx
		num++
	}

	p.record.Flags = flags
	p.record.Type = constType
	p.record.Group = constGroup
	p.record.InstanceEx = constInstEx

	count := len(p.record.Entries)
	p.header.EntryCount = uint32(count)
	headerSize := 4 * (1 + num)
	bodySize := count * 4 * (8 - num)
	p.header.RecordSize = uint32(headerSize + bodySize)

	p.header.RecordPosition = uint64(offset)

	return nil
}
