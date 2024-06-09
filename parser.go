package goson

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"
)

type Parser struct {
	json  string
	stack []byte
	top   int
}

func (p *Parser) push(b byte) {
	p.stack = append(p.stack, b)
	p.top++
}

func (p *Parser) pushs(bs []byte) {
	p.stack = append(p.stack, bs...)
	p.top += len(bs)
}

func (p *Parser) pop(size int) []byte {
	s := p.stack[p.top-size:]
	p.stack = p.stack[:p.top-size]
	p.top -= size
	return s
}

func (p *Parser) parseWhiteSpace() {
	for len(p.json) != 0 && (p.json[0] == ' ' || p.json[0] == '\t' || p.json[0] == '\n' || p.json[0] == '\r') {
		p.json = p.json[1:]
	}
}

func (p *Parser) parseLiteral(json string, t Type) (*Value, error) {
	var v Value
	if len(p.json) < len(json) || p.json[:len(json)] != json {
		return &v, ErrParseInvalidValue
	}
	p.json = p.json[len(json):]
	v.t = t
	return &v, nil
}

func isDigit1To9(b byte) bool {
	return b >= '1' && b <= '9'
}

func isDigit(b byte) bool {
	return b == '0' || isDigit1To9(b)
}

func (p *Parser) parseNumber() (*Value, error) {
	var v Value
	i := 0

	if i < len(p.json) && p.json[i] == '-' {
		i++
	}
	if i < len(p.json) && p.json[i] == '0' {
		i++
	} else {
		if i >= len(p.json) || !isDigit1To9(p.json[i]) {
			return &v, ErrParseInvalidValue
		}
		for i++; i < len(p.json) && isDigit(p.json[i]); i++ {
		}
	}
	if i < len(p.json) && p.json[i] == '.' {
		i++
		if i >= len(p.json) || !isDigit(p.json[i]) {
			return &v, ErrParseInvalidValue
		}
		for i++; i < len(p.json) && isDigit(p.json[i]); i++ {
		}
	}
	if i < len(p.json) && (p.json[i] == 'e' || p.json[i] == 'E') {
		i++
		if i < len(p.json) && (p.json[i] == '+' || p.json[i] == '-') {
			i++
		}
		if i >= len(p.json) || !isDigit(p.json[i]) {
			return &v, ErrParseInvalidValue
		}
		for i++; i < len(p.json) && isDigit(p.json[i]); i++ {
		}
	}

	n, err := strconv.ParseFloat(p.json[:i], 64)
	if err != nil && errors.Is(err, strconv.ErrRange) {
		return &v, ErrParseNumberTooBig
	}

	v.t = NUMBER
	v.n = n
	p.json = p.json[i:]
	return &v, nil
}

func (p *Parser) parseHex4() (uint32, error) {
	var u uint32
	for i := 0; i < 4; i++ {
		if len(p.json) == 0 {
			return 0, ErrParseInvalidUnicodeHex
		}

		ch := p.json[0]
		p.json = p.json[1:]
		u <<= 4

		switch {
		case ch >= '0' && ch <= '9':
			u |= uint32(ch - '0')
		case ch >= 'A' && ch <= 'F':
			u |= uint32(ch - 'A' + 10)
		case ch >= 'a' && ch <= 'f':
			u |= uint32(ch - 'a' + 10)
		default:
			return 0, ErrParseInvalidUnicodeHex
		}
	}
	return u, nil
}

func (p *Parser) encodeUTF8(u uint32) {
	buf := make([]byte, 4)
	n := utf8.EncodeRune(buf, rune(u))
	buf = buf[:n]
	p.pushs(buf)
}

func (p *Parser) parseString() (*Value, error) {
	var v Value
	var err error
	var s string
	if s, err = p.parseStringRaw(); err != nil {
		return &v, err
	}
	v.t = STRING
	v.s = s
	return &v, nil
}

func (p *Parser) parseStringRaw() (string, error) {
	if len(p.json) != 0 && p.json[0] != '"' {
		return "", fmt.Errorf(`missing close '"'`)
	}
	p.json = p.json[1:]
	if len(p.json) == 0 {
		return "", ErrParseMissQuotationMark
	}

	head := p.top

	for {
		ch := p.json[0]
		p.json = p.json[1:]
		if len(p.json) == 0 && ch != '"' {
			p.top = head
			return "", ErrParseMissQuotationMark
		}

		switch ch {
		case '"':
			l := p.top - head
			s := p.pop(l)
			return string(s), nil
		case '\\':
			ch = p.json[0]
			p.json = p.json[1:]
			switch ch {
			case '"':
				p.push('"')
			case '\\':
				p.push('\\')
			case '/':
				p.push('/')
			case 'b':
				p.push('\b')
			case 'f':
				p.push('\f')
			case 'n':
				p.push('\n')
			case 'r':
				p.push('\r')
			case 't':
				p.push('\t')
			case 'u':
				u, err := p.parseHex4()
				if err != nil {
					p.top = head
					return "", err
				}
				if u >= 0xD800 && u <= 0xDBFF {
					h := u

					ch = p.json[0]
					p.json = p.json[1:]
					if ch != '\\' {
						p.top = head
						return "", ErrParseInvalidUnicodeSurrogate
					}

					ch = p.json[0]
					p.json = p.json[1:]
					if ch != 'u' {
						p.top = head
						return "", ErrParseInvalidUnicodeSurrogate
					}

					u, err = p.parseHex4()
					if err != nil {
						p.top = head
						return "", ErrParseInvalidUnicodeSurrogate
					}

					if !(u >= 0xDC00 && u <= 0xDFFF) {
						p.top = head
						return "", ErrParseInvalidUnicodeSurrogate
					}

					l := u
					u = 0x10000 + (h-0xD800)*0x400 + (l - 0xDC00)
				}
				p.encodeUTF8(u)
			default:
				p.top = head
				return "", ErrParseInvalidStringEscape
			}
		default:
			if uint(ch) < 0x20 {
				p.top = head
				return "", ErrParseInvalidStringChar
			}
			p.push(ch)
		}
	}
}

func (p *Parser) parseArray() (*Value, error) {
	var v Value

	if len(p.json) == 0 || p.json[0] != '[' {
		return &v, fmt.Errorf(`missing close '['`)
	}
	p.json = p.json[1:]

	p.parseWhiteSpace()

	if len(p.json) != 0 && p.json[0] == ']' {
		p.json = p.json[1:]
		v.t = ARRAY
		return &v, nil
	}

	var err error

	for {
		var e *Value
		if e, err = p.parseValue(); err != nil {
			break
		}

		v.a = append(v.a, e)

		p.parseWhiteSpace()
		if len(p.json) != 0 && p.json[0] == ',' {
			p.json = p.json[1:]
			p.parseWhiteSpace()
		} else if len(p.json) != 0 && p.json[0] == ']' {
			p.json = p.json[1:]
			v.t = ARRAY
			return &v, nil
		} else {
			err = ErrParseMissCommaOrSquareBracket
			break
		}
	}

	v.a = make([]*Value, 0)

	return &v, err
}

func (p *Parser) parseObject() (*Value, error) {
	var v Value

	if len(p.json) == 0 || p.json[0] != '{' {
		return &v, fmt.Errorf(`missing close '{'`)
	}
	p.json = p.json[1:]

	p.parseWhiteSpace()

	if len(p.json) != 0 && p.json[0] == '}' {
		p.json = p.json[1:]
		v.t = OBJECT
		return &v, nil
	}

	var err error

	for {
		var kv KV
		var s string
		var vv *Value
		if len(p.json) == 0 || p.json[0] != '"' {
			err = ErrParseMissKey
			break
		}
		if s, err = p.parseStringRaw(); err != nil {
			break
		}
		kv.k = s

		p.parseWhiteSpace()
		if len(p.json) == 0 || p.json[0] != ':' {
			err = ErrParseMissColon
			break
		}
		p.json = p.json[1:]
		p.parseWhiteSpace()

		if vv, err = p.parseValue(); err != nil {
			break
		}
		kv.v = vv

		v.o = append(v.o, &kv)

		p.parseWhiteSpace()
		if len(p.json) != 0 && p.json[0] == ',' {
			p.json = p.json[1:]
			p.parseWhiteSpace()
		} else if len(p.json) != 0 && p.json[0] == '}' {
			p.json = p.json[1:]
			v.t = OBJECT
			return &v, nil
		} else {
			err = ErrParseMissCommaOrCurlyBracket
			break
		}
	}

	v.o = make([]*KV, 0)

	return &v, err
}

func (p *Parser) parseValue() (*Value, error) {
	var v Value
	if len(p.json) == 0 {
		return &v, nil
	}
	switch p.json[0] {
	case 't':
		return p.parseLiteral("true", TRUE)
	case 'f':
		return p.parseLiteral("false", FALSE)
	case 'n':
		return p.parseLiteral("null", NULL)
	case '"':
		return p.parseString()
	case '[':
		return p.parseArray()
	case '{':
		return p.parseObject()
	default:
		return p.parseNumber()
	}
}

func (p *Parser) Parse(s string) (*Value, error) {
	p.json = s
	p.parseWhiteSpace()

	var err error
	var v *Value
	if v, err = p.parseValue(); err == nil {
		p.parseWhiteSpace()
		if len(p.json) != 0 {
			return &Value{}, ErrParseRootNotSingular
		}
	}

	if p.top != 0 {
		panic("stack is not empty")
	}
	p.stack = []byte{}

	return v, err
}
