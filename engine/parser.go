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

package engine

import (
	"fmt"
)

type parser struct {
	bytes []byte
	curr  int
	line  int
	char  int
}

func (p *parser) panic(message string, args ...interface{}) {
	m := fmt.Sprintf(message, args...)
	panic(fmt.Errorf("parse (%v, %v): %v", p.line+1, p.char+1, m))
}

func (p *parser) hasMore() bool {
	return p.curr < len(p.bytes)
}

func (p *parser) next() uint8 {
	return p.bytes[p.curr]
}

func (p *parser) trim() {
	for p.hasMore() {
		switch p.bytes[p.curr] {
		case ' ', '\t':
			p.curr++
			p.char++
		default:
			return
		}
	}
}

func (p *parser) findCharOrEnd(c uint8) int {
	for i := p.curr; i < len(p.bytes); i++ {
		switch p.bytes[i] {
		case c, '\n':
			return i
		}
	}
	return len(p.bytes)
}

func (p *parser) findEndOfGroup(start, end uint8) (int, int, int) {
	n := 1
	l := p.line
	c := p.char
	for i := p.curr; i < len(p.bytes); i++ {
		switch p.bytes[i] {
		case '\n':
			l++
			c = 0
		case start:
			n++
			c++
		case end:
			n--
			if n == 0 {
				c++
				return i, l, c
			}
		default:
			c++
		}
	}
	p.panic("could not find matching '%v' to '%v'", end, start)
	return 0, 0, 0
}

func (p *parser) assumeWord(bytes []byte) {
	for _, b := range bytes {
		switch b {
		case '.', '"':
			p.panic("illegal characer '%v' found in word", rune(b))
		}
	}
}

func (p *parser) assumeName(bytes []byte) {
	for _, b := range bytes {
		switch b {
		case '"':
			p.panic("illegal characer '%v' found in name", rune(b))
		}
	}
}

func (p *parser) word() string {
	n := p.findCharOrEnd(' ')
	bytes := p.bytes[p.curr:n]
	p.assumeWord(bytes)

	p.curr = n
	p.char += len(bytes)
	p.trim()

	return string(bytes)
}

func (p *parser) name() string {
	n := p.findCharOrEnd(' ')
	bytes := p.bytes[p.curr:n]
	p.assumeName(bytes)

	p.curr = n
	p.char += len(bytes)
	p.trim()

	return string(bytes)
}

func (p *parser) string() string {
	p.curr++
	p.char++

	n := p.findCharOrEnd('"')
	bytes := p.bytes[p.curr:n]

	p.curr = n + 1
	p.char += len(bytes) + 1
	p.trim()

	return string(bytes)
}

func (p *parser) group(start, end uint8) *parser {
	p.curr++
	p.char++

	n, l, c := p.findEndOfGroup(start, end)

	parser := new(parser)
	parser.bytes = p.bytes[p.curr:n]
	parser.line = p.line
	parser.char = p.char
	parser.trim()

	p.curr = n + 1
	p.line = l
	p.char = c + 1
	p.trim()

	return parser
}

func (p *parser) ensure(word string) {
	w := p.word()
	if word != w {
		p.panic("expected %v but found %v", word, w)
	}
}

func (p *parser) end() {
	if p.curr >= len(p.bytes) {
		return
	}
	if p.bytes[p.curr] != '\n' {
		p.panic("expected new line")
	}
	p.curr++
	p.line++
	p.char = 0
	p.trim()
}
