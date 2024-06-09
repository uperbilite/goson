package goson

import "errors"

func isEqual(lhs *Value, rhs *Value) bool {
	if lhs.t != rhs.t {
		return false
	}
	switch lhs.t {
	case STRING:
		return lhs.s == rhs.s
	case NUMBER:
		return lhs.n == rhs.n
	case ARRAY:
		if len(lhs.a) != len(rhs.a) {
			return false
		}
		for i := 0; i < len(lhs.a); i++ {
			if !isEqual(lhs.a[i], rhs.a[i]) {
				return false
			}
		}
		return true
	case OBJECT:
		if len(lhs.o) != len(rhs.o) {
			return false
		}
		for i := 0; i < len(lhs.o); i++ {
			lv, err := lhs.getObjectValue(lhs.o[i].k)
			if errors.Is(err, ErrKeyNotExist) {
				return false
			}
			rv, err := rhs.getObjectValue(lhs.o[i].k)
			if errors.Is(err, ErrKeyNotExist) {
				return false
			}
			if !isEqual(lv, rv) {
				return false
			}
		}
		return true
	default:
		return true
	}
}
