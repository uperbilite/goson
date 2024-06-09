package goson

import (
	"fmt"
)

type Value struct {
	s string
	n float64
	a []*Value
	o []*KV
	t Type
}

type KV struct {
	k string
	v *Value
}

func (v *Value) setNull() {
	v.s = ""
	v.t = NULL
}

func (v *Value) free() {
	switch v.t {
	case STRING:
		v.s = ""
	case ARRAY:
		for _, e := range v.a {
			e.free()
		}
		v.a = make([]*Value, 0, 0)
	case OBJECT:
		for _, kv := range v.o {
			kv.k = ""
			kv.v.free()
		}
		v.o = make([]*KV, 0, 0)
	default:
	}
	v.setNull()
}

func (v *Value) getType() Type {
	return v.t
}

func (v *Value) setBoolean(b bool) {
	v.free()
	if b {
		v.t = TRUE
	} else {
		v.t = FALSE
	}
}

func (v *Value) getBoolean() (bool, error) {
	if v.t != TRUE && v.t != FALSE {
		return false, fmt.Errorf("value type is not boolean")
	}
	return v.t == TRUE, nil
}

func (v *Value) setNumber(n float64) {
	v.free()
	v.t = NUMBER
	v.n = n
}

func (v *Value) getNumber() (float64, error) {
	if v.t != NUMBER {
		return 0, fmt.Errorf("value type is not number")
	}
	return v.n, nil
}

func (v *Value) setString(s string) {
	v.free()
	v.t = STRING
	v.s = s
}

func (v *Value) getString() (string, error) {
	if v.t != STRING {
		return "", fmt.Errorf("value type is not string")
	}
	return v.s, nil
}

func (v *Value) setArray(size int) {
	v.free()
	v.t = ARRAY
	v.a = make([]*Value, 0, size)
}

func (v *Value) getArrayElement(index int) (*Value, error) {
	if v.t != ARRAY {
		return &Value{}, fmt.Errorf("value type is not array")
	}
	if index >= len(v.a) {
		return &Value{}, fmt.Errorf("index out of range")
	}
	return v.a[index], nil
}

func (v *Value) insertArrayElement(e *Value, index int) error {
	if v.t != ARRAY {
		return fmt.Errorf("value type is not array")
	}
	if index > len(v.a) {
		return fmt.Errorf("index out of range")
	}
	v.a = append(v.a, e)
	copy(v.a[index+1:], v.a[index:])
	v.a[index] = e
	return nil
}

func (v *Value) eraseArrayElement(index, count int) error {
	if v.t != ARRAY {
		return fmt.Errorf("value type is not array")
	}
	if index+count > len(v.a) {
		return fmt.Errorf("index out of range")
	}
	if count == 0 {
		return nil
	}
	v.a = append(v.a[:index], v.a[index+count:]...)
	return nil
}

func (v *Value) clearArray() error {
	return v.eraseArrayElement(0, len(v.a))
}

func (v *Value) setObject(size int) {
	v.free()
	v.t = OBJECT
	v.o = make([]*KV, 0, size)
}

func (v *Value) getObjectValue(key string) (*Value, error) {
	if v.t != OBJECT {
		return &Value{}, fmt.Errorf("value type is not object")
	}
	for _, kv := range v.o {
		if kv.k == key {
			return kv.v, nil
		}
	}
	return &Value{}, ErrKeyNotExist
}

func (v *Value) setObjectValue(key string, value *Value) error {
	if v.t != OBJECT {
		return fmt.Errorf("value type is not object")
	}
	for _, kv := range v.o {
		if kv.k == key {
			kv.v = value
			return nil
		}
	}
	v.o = append(v.o, &KV{key, value})
	return nil
}

func (v *Value) removeObjectValue(key string) error {
	if v.t != OBJECT {
		return fmt.Errorf("value type is not object")
	}
	var index int
	for i, kv := range v.o {
		if kv.k == key {
			index = i
			break
		}
	}
	v.o = append(v.o[:index], v.o[index+1:]...)
	return nil
}

func (v *Value) clearObject() error {
	if v.t != OBJECT {
		return fmt.Errorf("value type is not object")
	}
	v.o = v.o[0:0]
	return nil
}

func stringifyString(s string) string {
	result := ""

	result += "\""
	for _, ch := range s {
		switch ch {
		case '"':
			result += "\\\""
		case '\\':
			result += "\\\\"
		case '\b':
			result += "\\b"
		case '\f':
			result += "\\f"
		case '\n':
			result += "\\n"
		case '\r':
			result += "\\r"
		case '\t':
			result += "\\t"
		default:
			if uint(ch) < 0x20 {
				buf := fmt.Sprintf("\\u%04X", ch)
				result += buf
			} else {
				result += string(ch)
			}
		}
	}
	result += "\""

	return result
}

func (v *Value) stringifyValue() string {
	s := ""
	switch v.t {
	case NULL:
		s += "null"
	case FALSE:
		s += "false"
	case TRUE:
		s += "true"
	case NUMBER:
		s += fmt.Sprintf("%.17g", v.n)
	case STRING:
		s += stringifyString(v.s)
	case ARRAY:
		s += "["
		for i, e := range v.a {
			if i > 0 {
				s += ","
			}
			s += e.stringifyValue()
		}
		s += "]"
	case OBJECT:
		s += "{"
		for i, kv := range v.o {
			if i > 0 {
				s += ","
			}
			s += stringifyString(kv.k)
			s += ":"
			s += kv.v.stringifyValue()
		}
		s += "}"
	default:
		panic("invalid value type")
	}
	return s
}

func (v *Value) copy() *Value {
	var result Value
	switch v.t {
	case FALSE:
		result.setBoolean(false)
	case TRUE:
		result.setBoolean(true)
	case STRING:
		result.setString(v.s)
	case NUMBER:
		result.setNumber(v.n)
	case ARRAY:
		result.setArray(len(v.a))
		for i := 0; i < len(v.a); i++ {
			_ = result.insertArrayElement(v.a[i].copy(), i)
		}
	case OBJECT:
		result.setObject(len(v.o))
		for i := 0; i < len(v.o); i++ {
			value := v.o[i].v.copy()
			_ = result.setObjectValue(v.o[i].k, value)
		}
	default:
	}

	return &result
}
