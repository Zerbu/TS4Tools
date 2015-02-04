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
	"encoding/binary"
	"io/ioutil"
	"os"
)

const (
	unused                  = 3
	constantType            = 1 << 0
	constantGroup           = 1 << 1
	constantInstanceEx      = 1 << 2
	extendedCompressionType = 1 << (32 - 1)
)

var (
	identifier = [4]byte{'D', 'B', 'P', 'F'}
	headerSize = uint64(binary.Size(header{}))
)

func (p *Package) initFile(version Version) {
	p.header.Identifier = identifier
	p.header.FileVersion = version
	p.header.Unused = unused
	p.header.RecordPosition = headerSize
	p.record.Flags = constantType | constantGroup | constantInstanceEx
}

func (p *Package) readFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	p.file = file
	err = binary.Read(file, binary.LittleEndian, &p.header)
	if err != nil {
		return err
	}
	err = p.readRecord(file)
	if err != nil {
		return err
	}
	return nil
}

func (p *Package) writeFile(path string) error {

	temp, err := ioutil.TempFile(".", "dbpf-")
	if err != nil {
		return err
	}
	err = binary.Write(temp, binary.LittleEndian, &p.header)
	if err != nil {
		return err
	}
	err = p.writeRecord(temp)
	if err != nil {
		return err
	}
	temp.Close()
	p.file.Close()
	err = os.RemoveAll(path)
	if err != nil {
		file, _ := os.Open(p.file.Name())
		p.file = file
		return err
	}
	err = os.Rename(temp.Name(), path)
	if err != nil {
		file, _ := os.Open(p.file.Name())
		p.file = file
		return err
	}
	file, err := os.Open(path)
	p.file = file
	return err
}

func (p *Package) readRecord(file *os.File) error {
	_, err := file.Seek(int64(p.header.RecordPosition), os.SEEK_SET)
	if err != nil {
		return err
	}
	var flags uint32
	err = binary.Read(file, binary.LittleEndian, &flags)
	if err != nil {
		return err
	}
	p.record.Flags = flags
	if (flags & constantType) != 0 {
		err = binary.Read(file, binary.LittleEndian, &p.record.Type)
		if err != nil {
			return err
		}
	}
	if (flags & constantGroup) != 0 {
		err = binary.Read(file, binary.LittleEndian, &p.record.Group)
		if err != nil {
			return err
		}
	}
	if (flags & constantInstanceEx) != 0 {
		err = binary.Read(file, binary.LittleEndian, &p.record.InstanceEx)
		if err != nil {
			return err
		}
	}
	p.record.Entries = make([]*entry, p.header.EntryCount)
	for i := range p.record.Entries {
		entry, err := p.readEntry(file, flags)
		if err != nil {
			return err
		}
		p.record.Entries[i] = entry
	}
	return nil
}

func (p *Package) writeRecord(file *os.File) error {
	_, err := file.Seek(int64(p.header.RecordPosition), os.SEEK_SET)
	if err != nil {
		return err
	}
	flags := p.record.Flags
	err = binary.Write(file, binary.LittleEndian, &flags)
	if err != nil {
		return err
	}
	if (flags & constantType) != 0 {
		err = binary.Write(file, binary.LittleEndian, &p.record.Type)
		if err != nil {
			return err
		}
	}
	if (flags & constantGroup) != 0 {
		err = binary.Write(file, binary.LittleEndian, &p.record.Group)
		if err != nil {
			return err
		}
	}
	if (flags & constantInstanceEx) != 0 {
		err = binary.Write(file, binary.LittleEndian, &p.record.InstanceEx)
		if err != nil {
			return err
		}
	}
	for _, entry := range p.record.Entries {
		err = p.writeEntry(file, flags, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Package) readEntry(file *os.File, flags uint32) (*entry, error) {
	var entry entry
	if (flags & constantType) == 0 {
		err := binary.Read(file, binary.LittleEndian, &entry.Type)
		if err != nil {
			return nil, err
		}
	}
	if (flags & constantGroup) == 0 {
		err := binary.Read(file, binary.LittleEndian, &entry.Group)
		if err != nil {
			return nil, err
		}
	}
	if (flags & constantInstanceEx) == 0 {
		err := binary.Read(file, binary.LittleEndian, &entry.InstanceEx)
		if err != nil {
			return nil, err
		}
	}
	err := binary.Read(file, binary.LittleEndian, &entry.Fixed)
	if err != nil {
		return nil, err
	}
	if (entry.Fixed.CompressedSize & extendedCompressionType) != 0 {
		err = binary.Read(file, binary.LittleEndian, &entry.Extended)
		if err != nil {
			return nil, err
		}
	}
	return &entry, nil
}

func (p *Package) writeEntry(file *os.File, flags uint32, entry *entry) error {
	if (flags & constantType) == 0 {
		err := binary.Write(file, binary.LittleEndian, &entry.Type)
		if err != nil {
			return err
		}
	}
	if (flags & constantGroup) == 0 {
		err := binary.Write(file, binary.LittleEndian, &entry.Group)
		if err != nil {
			return err
		}
	}
	if (flags & constantInstanceEx) == 0 {
		err := binary.Write(file, binary.LittleEndian, &entry.InstanceEx)
		if err != nil {
			return err
		}
	}
	err := binary.Write(file, binary.LittleEndian, &entry.Fixed)
	if err != nil {
		return err
	}
	if (entry.Fixed.CompressedSize & extendedCompressionType) != 0 {
		err = binary.Write(file, binary.LittleEndian, &entry.Extended)
		if err != nil {
			return err
		}
	}
	return nil
}
